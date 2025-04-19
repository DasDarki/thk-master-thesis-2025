import seaborn as sns
import matplotlib.pyplot as plt
import numpy as np
from assets import get_error_free_data, get_output_stats_path, get_output_plot_path

def run_spearman_correlation():
    df = get_error_free_data()

    df_numeric = df.select_dtypes(include=['float64', 'int64']).copy()

    exclude = ['Id', 'ClientId']
    df_numeric = df_numeric[[col for col in df_numeric.columns if col not in exclude]]

    corr_matrix = df_numeric.corr(method='spearman')

    output_csv = get_output_stats_path("spearman_correlation_matrix.csv")
    corr_matrix.to_csv(output_csv, sep=';')
    print(f"Spearman-Korrelationsmatrix gespeichert unter {output_csv}")

    plt.subplots(figsize=(12, 10))
    mask = np.triu(np.ones_like(corr_matrix, dtype=bool))
    sns.heatmap(corr_matrix, annot=True, fmt=".2f", cmap="coolwarm", mask=mask,
                square=True, linewidths=0.5, cbar_kws={"label": "Spearman ρ"})
    plt.title("Spearman-Korrelationsmatrix (numerische Metriken)")
    plt.tight_layout()

    output_plot = get_output_plot_path("spearman_correlation_heatmap.png")
    plt.savefig(output_plot)
    plt.close()
    print(f"Heatmap gespeichert unter {output_plot}")