using System.Diagnostics;

namespace TestSuite;

/// <summary>
/// The test client is the interface for the test suite to the client.
/// </summary>
public class TestClient(string protocol)
{
    /// <summary>
    /// The protocol of the client.
    /// </summary>
    public string Protocol => protocol;
    
    /// <summary>
    /// Runs the test client and waits for the client to finish.
    /// </summary>
    /// <param name="id">The collector ID of the run.</param>
    /// <param name="local">Whether the run is local or not.</param>
    public void Run(int id, bool local)
    {
        var process = new Process
        {
            StartInfo = CreateRunProcess(id, local),
            EnableRaisingEvents = true,
        };

        process.Start();
        process.WaitForExit();
    }

    protected virtual ProcessStartInfo CreateRunProcess(int id, bool local)
    {
        var dir = GetTestClientWorkingDir();
        return new ProcessStartInfo
        {
            WorkingDirectory = dir,
            FileName = Path.Combine(dir, "app.exe"),
            Arguments = $"-r{id}{(local ? " -l" : "")}",
        };
    }

    protected string GetTestClientWorkingDir()
    {
        return Path.Combine(Environment.CurrentDirectory, "..", "..", "..", "..", protocol + "-client");
    }
}