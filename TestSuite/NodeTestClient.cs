using System.Diagnostics;

namespace TestSuite;

/// <summary>
/// Inherits <see cref="TestClient"/> and implements the test client for Node.
/// </summary>
public sealed class NodeTestClient(string protocol) : TestClient(protocol) 
{
    protected override ProcessStartInfo CreateRunProcess(int id, bool local)
    {
        return new ProcessStartInfo
        {
            WorkingDirectory = GetTestClientWorkingDir(),
            FileName = "node",
            Arguments = $"main.js -r{id}{(local ? " -l" : "")}",
        };
    }
}