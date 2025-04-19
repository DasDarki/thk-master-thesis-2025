import pandas as pd
import os

imported_cache = None
error_free_cache = None

def import_data():
    global imported_cache
    if imported_cache is not None:
        return imported_cache

    cwd = os.getcwd()
    f = os.path.join(cwd, '..', 'assets', 'results_clean.csv')
    df = pd.read_csv(f, sep=';')
    imported_cache = df
    return df

def get_error_free_data():
    global error_free_cache
    if error_free_cache is not None:
        return error_free_cache

    df = import_data()
    df_clean = df[df["Error"].isnull() | (df["Error"] == "")].copy()
    df_clean.drop(columns=["Error"], inplace=True)
    df_clean.reset_index(drop=True, inplace=True)
    error_free_cache = df_clean
    return df_clean

def get_output_path():
    cwd = os.getcwd()
    output_path = os.path.join(cwd, '..', 'assets', 'output')
    if not os.path.exists(output_path):
        os.makedirs(output_path)
    return output_path

def get_output_plot_path(name):
    output_path = get_output_path()
    plot_path = os.path.join(output_path, 'plots')
    if not os.path.exists(plot_path):
        os.makedirs(plot_path)
    return os.path.join(plot_path, name)

def get_output_table_path(name):
    output_path = get_output_path()
    table_path = os.path.join(output_path, 'tables')
    if not os.path.exists(table_path):
        os.makedirs(table_path)
    return os.path.join(table_path, name)

def get_output_stats_path(name):
    output_path = get_output_path()
    stats_path = os.path.join(output_path, 'stats')
    if not os.path.exists(stats_path):
        os.makedirs(stats_path)
    return os.path.join(stats_path, name)