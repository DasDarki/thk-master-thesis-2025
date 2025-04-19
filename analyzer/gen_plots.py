import os.path

import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
from assets import get_output_plot_path, get_error_free_data, import_data, get_output_stats_path


def annotate_bars(ax, fmt="{:.2f}", offset=3, rotation=0, fontsize=10):
    for container in ax.containers:
        for bar in container:
            height = bar.get_height()
            ax.annotate(fmt.format(height),
                        xy=(bar.get_x() + bar.get_width() / 2, height),
                        xytext=(0, offset),
                        textcoords="offset points",
                        ha='center', va='bottom',
                        rotation=rotation,
                        fontsize=fontsize)

def annotate_points(ax, x_col, y_col, data, fmt="{:.1f}"):
    for _, row in data.iterrows():
        ax.annotate(fmt.format(row[y_col]),
                    xy=(row[x_col], row[y_col]),
                    xytext=(0, 5),
                    textcoords="offset points",
                    ha="center", va="bottom",
                    fontsize=8)


def plot_performance_comparisons():
    df = get_error_free_data()

    # --- 1. Balken: Durchschnittlicher Durchsatz ---
    throughput_means = df.groupby("Protocol")["ThroughputMbps"].mean().sort_values()
    fig1, ax1 = plt.subplots(figsize=(8, 5))
    throughput_means.plot(kind="bar", color="cornflowerblue", ax=ax1)
    plt.title("Average Throughput per Protocol")
    plt.ylabel("Throughput (Mbps)")
    plt.xlabel("Protocol")
    plt.grid(axis="y", linestyle="--", alpha=0.6)
    annotate_bars(ax1)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("throughput_mean_per_protocol.png"))
    plt.close(fig1)

    # --- 2. Balken: Durchschnittliche Bandbreiteneffizienz ---
    # Rausgeflogen weil: Effizient ziemlich gleich

    # --- 3. Boxplot: Verteilung des Durchsatzes ---
    fig3 = plt.figure(figsize=(10, 6))
    sns.boxplot(x="Protocol", y="ThroughputMbps", data=df, palette="Blues")
    plt.title("Throughput Distribution per Protocol")
    plt.ylabel("Throughput (Mbps)")
    plt.grid(axis="y", linestyle="--", alpha=0.6)
    plt.tight_layout()
    fig3.savefig(get_output_plot_path("throughput_boxplot_per_protocol.png"))
    plt.close(fig3)

    # --- 4. Boxplot: Verteilung der Effizienz ---
    fig4 = plt.figure(figsize=(10, 6))
    sns.boxplot(x="Protocol", y="BandwidthEfficiency", data=df, palette="Greens")
    plt.title("Bandwidth Efficiency Distribution per Protocol")
    plt.ylabel("Efficiency (0–1)")
    plt.grid(axis="y", linestyle="--", alpha=0.6)
    plt.tight_layout()
    fig4.savefig(get_output_plot_path("bandwidth_efficiency_boxplot_per_protocol.png"))
    plt.close(fig4)

    # --- 5. Balken: Wer ist am schnellsten? (niedrigste TransferDuration) ---
    duration_means = df.groupby("Protocol")["TransferDuration"].mean().sort_values()
    fig5, ax5 = plt.subplots(figsize=(8, 5))
    duration_means.plot(kind="bar", color="salmon", ax=ax5)
    plt.title("Average Transfer Duration per Protocol")
    plt.ylabel("Duration (seconds)")
    plt.xlabel("Protocol")
    plt.grid(axis="y", linestyle="--", alpha=0.6)
    annotate_bars(ax5)
    plt.tight_layout()
    fig5.savefig(get_output_plot_path("transfer_duration_mean_per_protocol.png"))
    plt.close(fig5)

    # --- 6. Vergleich: Schnelligkeit vs. Verbindungsdauer vs. Fehlerquote ---
    transfer_duration = df.groupby("Protocol")["TransferDuration"].mean()
    connection_duration = df.groupby("Protocol")["ConnectionDuration"].mean()
    total_counts = df.groupby("Protocol").size()
    original = pd.DataFrame(import_data())
    error_counts = original[original["Error"].notnull()].groupby("Protocol").size()
    error_rate = (error_counts / total_counts).fillna(0)

    comparison_df = pd.DataFrame({
        "Transfer Duration (s)": transfer_duration,
        "Connection Duration (s)": connection_duration,
        "Error Rate (%)": error_rate * 100
    }).round(2)

    fig6, ax6 = plt.subplots(figsize=(10, 6))
    comparison_df.plot(kind="bar", ax=ax6, color=["salmon", "orange", "gray"])
    plt.title("Speed vs Stability per Protocol")
    plt.ylabel("Value")
    plt.xlabel("Protocol")
    plt.grid(axis="y", linestyle="--", alpha=0.5)
    annotate_bars(ax6, fontsize=8, offset=3)
    plt.tight_layout()
    fig6.savefig(get_output_plot_path("transfer_speed_vs_costs_per_protocol.png"))
    plt.close(fig6)

    print("Performance comparison plots saved.")


def plot_scalability():
    df = get_error_free_data()
    grouped = df.groupby(["Protocol", "ParallelClients"])

    # --- Plot 1: Throughput vs Clients ---
    throughput = grouped["ThroughputMbps"].mean().reset_index()
    fig1, ax1 = plt.subplots(figsize=(10, 6))
    sns.lineplot(data=throughput, x="ParallelClients", y="ThroughputMbps", hue="Protocol", marker="o", ax=ax1)
    for proto in throughput["Protocol"].unique():
        annotate_points(ax1, "ParallelClients", "ThroughputMbps", throughput[throughput["Protocol"] == proto])
    plt.title("Scalability: Throughput vs Parallel Clients")
    plt.ylabel("Throughput (Mbps)")
    plt.grid(True, linestyle="--", alpha=0.6)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("scalability_throughput_vs_clients.png"))
    plt.close(fig1)

    # --- Plot 2: CPU Client Load (Before/While/After) ---
    for phase in ["Before", "While", "After"]:
        col = f"CpuClient{phase}"
        cpu = grouped[col].mean().reset_index()
        fig, ax = plt.subplots(figsize=(10, 6))
        sns.lineplot(data=cpu, x="ParallelClients", y=col, hue="Protocol", marker="o", ax=ax)
        for proto in cpu["Protocol"].unique():
            annotate_points(ax, "ParallelClients", col, cpu[cpu["Protocol"] == proto])
        plt.title(f"Scalability: CPU Client {phase} vs Parallel Clients")
        plt.ylabel("CPU Usage (%)")
        plt.grid(True, linestyle="--", alpha=0.6)
        plt.tight_layout()
        fig.savefig(get_output_plot_path(f"scalability_cpu_client_{phase.lower()}.png"))
        plt.close(fig)

    # --- Plot 3: RAM Client Load (Before/While/After) ---
    for phase in ["Before", "While", "After"]:
        col = f"RamClient{phase}"
        ram = grouped[col].mean().reset_index()
        fig, ax = plt.subplots(figsize=(10, 6))
        sns.lineplot(data=ram, x="ParallelClients", y=col, hue="Protocol", marker="o", ax=ax)
        for proto in ram["Protocol"].unique():
            annotate_points(ax, "ParallelClients", col, ram[ram["Protocol"] == proto])
        plt.title(f"Scalability: RAM Client {phase} vs Parallel Clients")
        plt.ylabel("RAM Usage (Bytes)")
        plt.grid(True, linestyle="--", alpha=0.6)
        plt.tight_layout()
        fig.savefig(get_output_plot_path(f"scalability_ram_client_{phase.lower()}.png"))
        plt.close(fig)

    print("Scalability plots saved.")
def plot_resource_flow():
    df = get_error_free_data()

    protocols = df["Protocol"].unique()
    phases = ["Before", "While", "After"]

    # CPU: Mittelwerte vorbereiten
    cpu_data = {
        proto: [
            df[df["Protocol"] == proto][f"CpuClient{phase}"].mean()
            for phase in phases
        ] for proto in protocols
    }

    # RAM: Mittelwerte vorbereiten
    ram_data = {
        proto: [
            df[df["Protocol"] == proto][f"RamClient{phase}"].mean()
            for phase in phases
        ] for proto in protocols
    }

    # --- CPU Plot ---
    fig1, ax1 = plt.subplots(figsize=(10, 6))
    for proto, values in cpu_data.items():
        ax1.plot(phases, values, marker="o", label=proto)
        for i, v in enumerate(values):
            ax1.annotate(f"{v:.1f}", (phases[i], v), textcoords="offset points", xytext=(0, 5), ha='center', fontsize=8)
    ax1.set_title("Client CPU Usage Flow by Protocol")
    ax1.set_ylabel("CPU Usage (%)")
    ax1.set_xlabel("Phase")
    ax1.grid(True, linestyle="--", alpha=0.6)
    ax1.legend()
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("cpu_flow_per_protocol.png"))
    plt.close(fig1)

    # --- RAM Plot ---
    fig2, ax2 = plt.subplots(figsize=(10, 6))
    for proto, values in ram_data.items():
        ax2.plot(phases, values, marker="o", label=proto)
        for i, v in enumerate(values):
            ax2.annotate(f"{v:.0f}", (phases[i], v), textcoords="offset points", xytext=(0, 5), ha='center', fontsize=8)
    ax2.set_title("Client RAM Usage Flow by Protocol")
    ax2.set_ylabel("RAM Usage (Bytes)")
    ax2.set_xlabel("Phase")
    ax2.grid(True, linestyle="--", alpha=0.6)
    ax2.legend()
    plt.tight_layout()
    fig2.savefig(get_output_plot_path("ram_flow_per_protocol.png"))
    plt.close(fig2)

    print("Resource flow plots saved.")

def plot_robustness():
    df_ok = get_error_free_data()
    df_all = pd.DataFrame(import_data())

    # --- 1. Fehlerquote pro Protocol + ParallelClients ---
    error_counts = df_all[df_all["Error"].notnull()].groupby(["Protocol", "ParallelClients"]).size()
    total_counts = df_all.groupby(["Protocol", "ParallelClients"]).size()
    error_rate = (error_counts / total_counts).fillna(0) * 100

    error_df = error_rate.reset_index(name="ErrorRate")

    fig1, ax1 = plt.subplots(figsize=(10, 6))
    sns.lineplot(data=error_df, x="ParallelClients", y="ErrorRate", hue="Protocol", marker="o", ax=ax1)
    for _, row in error_df.iterrows():
        ax1.annotate(f"{row['ErrorRate']:.1f}%", (row["ParallelClients"], row["ErrorRate"]),
                     textcoords="offset points", xytext=(0, 5), ha='center', fontsize=8)
    ax1.set_title("Error Rate per Protocol and Parallel Clients")
    ax1.set_ylabel("Error Rate (%)")
    ax1.set_xlabel("Parallel Clients")
    ax1.grid(True, linestyle="--", alpha=0.6)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("robustness_errorrate_per_protocol_clients.png"))
    plt.close(fig1)

    # --- 2. Boxplot TransferDuration bei ParallelClients = 20 ---
    df_20 = df_ok[df_ok["ParallelClients"] == 20]

    fig2 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df_20, x="Protocol", y="TransferDuration", palette="Set2")
    plt.title("Transfer Duration Distribution at 20 Parallel Clients")
    plt.ylabel("Transfer Duration (s)")
    plt.grid(True, linestyle="--", alpha=0.6)
    plt.tight_layout()
    fig2.savefig(get_output_plot_path("robustness_duration_boxplot_clients20.png"))
    plt.close(fig2)

    print("Robustness plots saved.")

def plot_handshake_times():
    df = get_error_free_data()

    # Konvertiere Zeitspalten in Timestamps (falls nicht bereits)
    df["TestBegin"] = pd.to_datetime(df["TestBegin"]).dt.tz_localize(None)
    df["TransferStart"] = pd.to_datetime(df["TransferStart"], unit="s").dt.tz_localize(None)
    df["HandshakeTime"] = (df["TestBegin"] - df["TransferStart"]).dt.total_seconds()

    # --- 1. Balkenplot: Mittelwert HandshakeTime pro Protokoll ---
    handshake_means = df.groupby("Protocol")["HandshakeTime"].mean().sort_values()

    fig1, ax1 = plt.subplots(figsize=(8, 5))
    handshake_means.plot(kind="bar", color="mediumorchid", ax=ax1)
    for i, v in enumerate(handshake_means):
        ax1.annotate(f"{v:.2f}s", (i, v), xytext=(0, 5), textcoords="offset points", ha='center', fontsize=9)
    ax1.set_title("Average Handshake Time per Protocol")
    ax1.set_ylabel("Seconds")
    ax1.set_xlabel("Protocol")
    ax1.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("handshake_time_mean_per_protocol.png"))
    plt.close(fig1)

    # --- 2. Boxplot: Verteilung der Handshake-Zeiten ---
    fig2 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df, x="Protocol", y="HandshakeTime", palette="Purples")
    plt.title("Handshake Time Distribution per Protocol")
    plt.ylabel("Seconds")
    plt.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig2.savefig(get_output_plot_path("handshake_time_boxplot_per_protocol.png"))
    plt.close(fig2)

    print("Handshake time plots saved.")

def plot_timeslot_effects():
    df_ok = get_error_free_data()
    df_all = pd.DataFrame(import_data())

    # --- 1. Boxplot Throughput je TimeSlot ---
    fig1 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df_ok, x="TimeSlot", y="ThroughputMbps", palette="Blues")
    plt.title("Throughput per TimeSlot")
    plt.ylabel("Throughput (Mbps)")
    plt.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("timeslot_throughput_boxplot.png"))
    plt.close(fig1)

    # --- 2. Boxplot TransferDuration je TimeSlot ---
    fig2 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df_ok, x="TimeSlot", y="TransferDuration", palette="Oranges")
    plt.title("Transfer Duration per TimeSlot")
    plt.ylabel("Duration (s)")
    plt.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig2.savefig(get_output_plot_path("timeslot_duration_boxplot.png"))
    plt.close(fig2)

    # --- 3. Fehlerquote je TimeSlot ---
    error_counts = df_all[df_all["Error"].notnull()].groupby("TimeSlot").size()
    total_counts = df_all.groupby("TimeSlot").size()
    error_rate = (error_counts / total_counts).fillna(0) * 100
    error_df = error_rate.reset_index(name="ErrorRate")

    fig3, ax3 = plt.subplots(figsize=(8, 5))
    sns.barplot(data=error_df, x="TimeSlot", y="ErrorRate", palette="Reds", ax=ax3)
    for idx, row in error_df.iterrows():
        ax3.annotate(f"{row['ErrorRate']:.1f}%", (idx, row["ErrorRate"] + 0.5),
                     ha="center", fontsize=9)
    ax3.set_title("Error Rate per TimeSlot")
    ax3.set_ylabel("Error Rate (%)")
    ax3.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig3.savefig(get_output_plot_path("timeslot_errorrate.png"))
    plt.close(fig3)

    print("TimeSlot analysis plots saved.")

def plot_environment_comparison():
    df = get_error_free_data()

    # Handshake-Time berechnen
    df["TestBegin"] = pd.to_datetime(df["TestBegin"]).dt.tz_localize(None)
    df["TransferStart"] = pd.to_datetime(df["TransferStart"], unit="s").dt.tz_localize(None)
    df["HandshakeTime"] = (df["TestBegin"] - df["TransferStart"]).dt.total_seconds()

    # Plot 1: Boxplot Throughput
    fig1 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df, x="Environment", y="ThroughputMbps", palette="Blues")
    plt.title("Throughput by Environment")
    plt.ylabel("Throughput (Mbps)")
    plt.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig1.savefig(get_output_plot_path("env_throughput_boxplot.png"))
    plt.close(fig1)

    # Plot 2: Boxplot Transfer Duration
    fig2 = plt.figure(figsize=(10, 6))
    sns.boxplot(data=df, x="Environment", y="TransferDuration", palette="Oranges")
    plt.title("Transfer Duration by Environment")
    plt.ylabel("Duration (s)")
    plt.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig2.savefig(get_output_plot_path("env_transferduration_boxplot.png"))
    plt.close(fig2)

    # Plot 3: Mean HandshakeTime by Environment
    handshake_means = df.groupby("Environment")["HandshakeTime"].mean()
    fig3, ax3 = plt.subplots(figsize=(8, 5))
    handshake_means.plot(kind="bar", color="mediumseagreen", ax=ax3)
    for i, v in enumerate(handshake_means):
        ax3.annotate(f"{v:.2f}s", (i, v + 0.2), ha="center", fontsize=9)
    ax3.set_title("Average Handshake Time by Environment")
    ax3.set_ylabel("Seconds")
    ax3.grid(True, linestyle="--", alpha=0.5)
    plt.tight_layout()
    fig3.savefig(get_output_plot_path("env_handshake_mean.png"))
    plt.close(fig3)

    print("Environment comparison plots saved.")

def plot_dunn_heatmaps():
    # Lade alle Post-hoc-Dunn-Ergebnisse
    files = {
        "TransferDuration": get_output_stats_path(os.path.join('posthoc_dunn', 'posthoc_dunn_TransferDuration.csv')),
        "ThroughputMbps": get_output_stats_path(os.path.join('posthoc_dunn', 'posthoc_dunn_ThroughputMbps.csv')),
        "CpuClientWhile": get_output_stats_path(os.path.join('posthoc_dunn', 'posthoc_dunn_CpuClientWhile.csv')),
        "RamClientWhile": get_output_stats_path(os.path.join('posthoc_dunn', 'posthoc_dunn_RamClientWhile.csv'))
    }

    # Erzeuge Heatmaps
    for label, path in files.items():
        df = pd.read_csv(path, sep=";", index_col=0)
        plt.figure(figsize=(8, 6))
        ax = sns.heatmap(df, annot=True, fmt=".3f", cmap="coolwarm", linewidths=0.5, square=True,
                         cbar_kws={"label": "p-Wert"})
        plt.title(f"Post-hoc Dunn-Test: {label}")
        plt.tight_layout()
        out_path = get_output_plot_path(f"dunn_heatmap_{label}.png")
        plt.savefig(out_path)
        plt.close()