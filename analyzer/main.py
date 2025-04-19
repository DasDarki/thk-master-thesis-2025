from gen_correlation_stats import run_spearman_correlation
from gen_descriptive_stats import generate_descriptive_stats
from gen_kruskal_stats import generate_kruskal_stats
from gen_plots import plot_performance_comparisons, plot_scalability, plot_resource_flow, plot_robustness, \
    plot_handshake_times, plot_timeslot_effects, plot_environment_comparison, plot_dunn_heatmaps


def main():
    generate_descriptive_stats()
    generate_kruskal_stats()

    plot_performance_comparisons()
    plot_scalability()
    plot_resource_flow()
    plot_robustness()
    plot_handshake_times()
    plot_timeslot_effects()
    plot_environment_comparison()
    plot_dunn_heatmaps()

    run_spearman_correlation()

if __name__ == '__main__':
    main()
