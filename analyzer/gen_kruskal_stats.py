import pandas as pd
import scikit_posthocs as sp
import os
from scipy.stats import kruskal

from assets import get_error_free_data, get_output_stats_path


def run_kruskal_tests(df):
    test_metrics = {
        "TransferDuration": "Latenz (Transferdauer)",
        "ThroughputMbps": "Durchsatz",
        "BandwidthEfficiency": "Bandbreiteneffizienz",
        "CpuClientWhile": "CPU-Nutzung",
        "RamClientWhile": "RAM-Nutzung"
    }

    results = []

    # Konvertiere relevante Spalten in numerisch
    for col in test_metrics.keys():
        df[col] = pd.to_numeric(df[col], errors="coerce")

    for col, label in test_metrics.items():
        groups = [g[col].dropna().values for _, g in df.groupby("Protocol")]
        if all((group == group[0]).all() for group in groups if len(group) > 0):
            results.append({
                "Metrik": label,
                "Test": "Kruskal-Wallis",
                "Statistik": "n/a",
                "p-Wert": "n/a",
                "Anmerkung": "Keine Varianz in mind. 1 Gruppe"
            })
        else:
            stat, p = kruskal(*groups)
            results.append({
                "Metrik": label,
                "Test": "Kruskal-Wallis",
                "Statistik": round(stat, 4),
                "p-Wert": round(p, 6),
                "Anmerkung": "Signifikant" if p < 0.05 else "Nicht signifikant"
            })

    return pd.DataFrame(results)

def run_posthoc_dunn(df, output_dir):
    test_metrics = {
        "TransferDuration": "Latenz (Transferdauer)",
        "ThroughputMbps": "Durchsatz",
        "CpuClientWhile": "CPU-Nutzung",
        "RamClientWhile": "RAM-Nutzung"
    }

    os.makedirs(output_dir, exist_ok=True)

    for col, label in test_metrics.items():
        df[col] = pd.to_numeric(df[col], errors="coerce")
        df_subset = df[["Protocol", col]].dropna()

        # Wenn genug Varianz, führe Dunn-Test aus
        if df_subset[col].nunique() > 1 and df_subset["Protocol"].nunique() > 1:
            posthoc = sp.posthoc_dunn(df_subset, val_col=col, group_col="Protocol", p_adjust="bonferroni")
            output_file = os.path.join(output_dir, f"posthoc_dunn_{col}.csv")
            posthoc.to_csv(output_file, sep=';')
            print(f"Post-hoc Dunn-Test für '{label}' gespeichert unter {output_file}")
        else:
            print(f"Nicht genügend Varianz für Dunn-Test bei '{label}' – übersprungen.")

def generate_kruskal_stats():
    df = get_error_free_data()
    kruskal_results = run_kruskal_tests(df)
    output_path = get_output_stats_path('kruskal_results.csv')
    kruskal_results.to_csv(output_path, sep=';', index=False)
    print(f"Kruskal-Wallis test results saved to {output_path}")

    posthoc_path = os.path.join(os.path.dirname(output_path), "posthoc_dunn")
    run_posthoc_dunn(df, posthoc_path)

