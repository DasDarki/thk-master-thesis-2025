from assets import get_error_free_data, get_output_stats_path

def generate_descriptive_stats():
    exclude_columns = ['Id', 'ClientId', 'TestBegin', 'TestEnd', 'TransferStart', 'TransferEnd', 'ParallelClients']

    df = get_error_free_data()
    numeric_cols = df.select_dtypes(include='number').columns
    included_cols = [col for col in numeric_cols if col not in exclude_columns]

    grouped = df.groupby(["Protocol", "Environment", "ParallelClients", "TimeSlot"])
    summary_stats = grouped[included_cols].describe().round(2)
    summary_flat = summary_stats.reset_index()

    output_path = get_output_stats_path('descriptive_stats.csv')
    summary_flat.to_csv(output_path, sep=';', index=False)

    grouped2 = df.groupby(["Protocol"])
    summary_stats2 = grouped2[included_cols].describe().round(2)
    summary_flat2 = summary_stats2.reset_index()

    output_path2 = get_output_stats_path('descriptive_stats_protocol.csv')
    summary_flat2.to_csv(output_path2, sep=';', index=False)

    print(f"Descriptive statistics saved to {output_path}")