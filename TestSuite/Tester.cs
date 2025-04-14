using RestSharp;

namespace TestSuite;

/// <summary>
/// The runner/manager for all tests.
/// </summary>
public static class Tester
{
    #region Clients

    public static readonly TestClient Http3Client = new(ProtocolHTTP3);
    public static readonly TestClient WebTransportClient = new(ProtocolWebTransport);
    public static readonly TestClient WebSocketsClient = new(ProtocolWebSockets);
    public static readonly TestClient WebRTCClient = new NodeTestClient(ProtocolWebRTC);

    #endregion

    #region Runner

    /// <summary>
    /// Runs the test client and waits for the client to finish.
    /// </summary>
    public static void Run(TestClient client, bool local, string timeSlot, int parallelClients = 1)
    {
        var tasks = new List<Task>();
        var errors = new List<string>();
        var env = local ? EnvironmentLocal : EnvironmentRemote;
        
        for (var i = 0; i < parallelClients; i++)
        {
            var cid = i + 1;
            tasks.Add(Task.Run(() =>
            {
                try
                {
                    Console.WriteLine("Running client [#{0}] {1}...", cid, client.Protocol);
                    var runID = GetRunID(client.Protocol, env, timeSlot, cid, parallelClients);
                    Console.WriteLine("[#{0}] Run ID: {1}", cid, runID);
                    client.Run(runID, local);
                    Console.WriteLine("[#{0}.{1}] Finished running client {2}", cid, runID, client.Protocol);
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Error running client [#{cid}] {client.Protocol}: {ex.Message}");
                    errors.Add($"Error running client [#{cid}] {client.Protocol}: {ex.Message}");
                }
            }));
        }
        
        Task.WaitAll(tasks.ToArray());
        
        if (errors.Count > 0)
        {
            Console.WriteLine("Errors occurred while running clients:");
            foreach (var error in errors)
            {
                Console.WriteLine(error);
            }
        }
        else
        {
            Console.WriteLine("All clients finished successfully.");
        }
    }

    #endregion

    #region Collector
    
    private static readonly RestClient RestClient = new("https://thkm25_collect.nauri.io");
    
    private static int GetRunID(string protocol, string env, string timeSlot, int clientID, int parallelClients = 1)
    {
        var request = new RestRequest("/begin");
        request.AddHeader("X-API-Key", "thk_masterthesis_2025_hwtwswrtc");
        request.AddHeader("Content-Type", "application/json");
        request.AddJsonBody(new
        {
            Protocol = protocol,
            Environment = env,
            TimeSlot = timeSlot,
            ClientID = clientID,
            ParallelClients = parallelClients
        });
        
        var response = RestClient.Post(request);
        if (response.StatusCode != System.Net.HttpStatusCode.OK)
        {
            throw new Exception($"Failed to get run ID: {response.Content}");
        }
        
        var runID = response.Content;
        if (string.IsNullOrEmpty(runID))
        {
            throw new Exception("Run ID is empty");
        }
        
        if (!int.TryParse(runID, out var id))
        {
            throw new Exception($"Run ID is not a valid integer: {runID}");
        }
        
        return id;
    }

    #endregion
    
    #region Constants

    public const string ProtocolHTTP3 = "http3";
    public const string ProtocolWebTransport = "webtransport";
    public const string ProtocolWebSockets = "websockets";
    public const string ProtocolWebRTC = "webrtc";
    public const string EnvironmentLocal = "local";
    public const string EnvironmentRemote = "remote";
    public const string TimeSlotMorning = "morning";
    public const string TimeSlotAfternoon = "afternoon";
    public const string TimeSlotEvening = "evening";
    public const string TimeSlotNight = "night";

    #endregion
}