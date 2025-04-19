using System.Globalization;

namespace TestSuite;

/// <summary>
/// The cleaner takes in a result csv from the collector and cleans it up.
/// </summary>
public static class Cleaner
{
    /// <summary>
    /// Starts the cleaning process.
    /// </summary>
    public static void Clean()
    {
        var results = ReadInputResults();
        
        var type = typeof(Result);
        var fields = type.GetProperties();

        var header = fields.Select(field => field.Name).ToList();
        var lines = new List<string> { string.Join(";", header) };
        
        foreach (var result in results)
        {
            result.ApplyFixes();
            
            var line = fields.Select(field =>
            {
                var objValue = field.GetValue(result);
                string value;
                if (objValue is DateTime dateTime)
                {
                    value = dateTime.ToString("yyyy-MM-ddTHH:mm:ssZ");
                }
                else if (objValue is long longValue)
                {
                    value = longValue.ToString();
                }
                else if (objValue is double doubleValue)
                {
                    value = doubleValue.ToString(CultureInfo.InvariantCulture);
                }
                else if (objValue is string stringValue)
                {
                    value = stringValue;
                }
                else if (objValue is int intValue)
                {
                    value = intValue.ToString();
                }
                else
                {
                    value = objValue?.ToString() ?? "";
                }
                
                if (value.Contains(';'))
                {
                    value = "\"" + value + "\"";
                }
                
                return value;
            }).ToList();
            
            lines.Add(string.Join(";", line));
        }
        
        File.WriteAllLines(GetOutputCsv(), lines);
    }

    private static List<Result> ReadInputResults()
    {
        var results = new List<Result>();
        var lines = File.ReadAllLines(GetResultsCsv());
        var headers = lines[0].Split(';');
        for (var i = 1; i < lines.Length; i++)
        {
            var line = lines[i];
            var data = line.Split(';');
            var result = new Dictionary<string, string>();
            for (var j = 0; j < headers.Length; j++)
            {
                result[headers[j]] = data[j];
            }

            results.Add(Result.Parse(result));
        }
        
        return results;
    }

    private class Result
    {
        public int Id { get; set; }
        
        public string Protocol { get; set; }
        
        public string Environment { get; set; }
        
        public string TimeSlot { get; set; }
        
        public DateTime TestBegin { get; set; }
        
        public DateTime TestEnd { get; set; }
        
        public int ClientId { get; set; }
        
        public int ParallelClients { get; set; }
        
        public long TransferStart { get; set; }
        
        public long TransferEnd { get; set; }
        
        public long BytesPayload { get; set; }
        
        public double CpuClientBefore { get; set; }
        
        public double CpuClientAfter { get; set; }
        
        public double CpuClientWhile { get; set; }
        
        public double CpuServerBefore { get; set; }
        
        public double CpuServerAfter { get; set; }
        
        public double CpuServerWhile { get; set; }
        
        public long RamClientBefore { get; set; }
        
        public long RamClientAfter { get; set; }
        
        public long RamClientWhile { get; set; }
        
        public long RamServerBefore { get; set; }
        
        public long RamServerAfter { get; set; }
        
        public long RamServerWhile { get; set; }
        
        public long LostPackets { get; set; }
        
        public string Error { get; set; }
        
        public double ThroughputMbps => (float)BytesPayload / (TransferEnd - TransferStart) * 8 / 1000000;
        
        public long BytesSentTotal { get; set; }
        
        public double BandwidthEfficiency => (float)BytesPayload / BytesSentTotal;
        
        public long ConnectionDuration => (TestEnd - TestBegin).Ticks / TimeSpan.TicksPerSecond;
        
        public long TransferDuration => TransferEnd - TransferStart;

        public void ApplyFixes()
        {
            #region No Valid Test End

            // if TestEnd is 0001-01-01T00:00:00Z, set everything to 0 and Error to "ENDTIME NOT SET/COLLECTED"
            if (TestEnd.ToString("yyyy-MM-ddTHH:mm:ssZ") == "0001-01-01T00:00:00Z")
            {
                TransferStart = 0;
                TransferEnd = 0;
                BytesPayload = 0;
                CpuClientBefore = 0;
                CpuClientAfter = 0;
                CpuClientWhile = 0;
                CpuServerBefore = 0;
                CpuServerAfter = 0;
                CpuServerWhile = 0;
                RamClientBefore = 0;
                RamClientAfter = 0;
                RamClientWhile = 0;
                RamServerBefore = 0;
                RamServerAfter = 0;
                RamServerWhile = 0;
                LostPackets = 0;
                Error = "ENDTIME NOT SET/COLLECTED";
                return;
            }

            #endregion

            #region Restore any Test which has no TransferStart/TransferEnd but TestEnd/TestBegin

            if (TransferStart == 0 && TestBegin.ToString("yyyy-MM-ddTHH:mm:ssZ") != "0001-01-01T00:00:00Z")
            {
                TransferStart = GetSecondsSinceEpoch(TestBegin);
                Error = "TRANSFERSTART NOT SET/COLLECTED";
            }
            
            if (TransferEnd == 0 && TestEnd.ToString("yyyy-MM-ddTHH:mm:ssZ") != "0001-01-01T00:00:00Z")
            {
                TransferEnd = GetSecondsSinceEpoch(TestEnd);
                
                if (!string.IsNullOrEmpty(Error))
                {
                    Error += " / TRANSFEREND NOT SET/COLLECTED";
                }
                else
                {
                    Error = "TRANSFEREND NOT SET/COLLECTED";
                }
            }

            #endregion

            #region Fix Transfer End Times for WebRTC

            // if the protocol is "webrtc" the TransferStart uses milliseconds instead of seconds
            if (Protocol == Tester.ProtocolWebRTC)
            {
                TransferStart /= 1000;
            }

            #endregion

            #region Calculate/Approximate BytesSentTotal

            switch (Protocol)
            {
                case Tester.ProtocolHTTP3:
                case Tester.ProtocolWebTransport:
                {
                    const int chunkSize = 16 * 1024; // 16 KiB typical DATA frame size
                    const int overheadPerChunk = 60; // estimated average overhead (HTTP3 + QUIC + UDP + IP)

                    var chunkCount = (BytesPayload + chunkSize - 1) / chunkSize;
                    
                    BytesSentTotal = BytesPayload + chunkCount * overheadPerChunk;
                }
                    break;
                case Tester.ProtocolWebSockets:
                {
                    var chunks = (int)Math.Ceiling(BytesPayload / (64 * 1024f));
                    BytesSentTotal = BytesPayload + chunks * 8;
                }
                    break;
                case Tester.ProtocolWebRTC:
                {
                    var chunks = (int)Math.Ceiling(BytesPayload / (64 * 1024f));
                    BytesSentTotal = BytesPayload + chunks * 240; 
                    /*
                    A 64 KiB payload sent over WebRTC DataChannels is typically split into ~4 chunks of ~16 KiB each.
                    Each chunk carries around 60 bytes of protocol overhead (SCTP + DTLS + UDP + IP).
                    This means that a 64 KiB payload will be sent as 4 chunks of 16 KiB each, with a total overhead of ~240 bytes.
                    */
                }
                    break;
            }

            #endregion

            #region Clean Up Errors

            if (Error == "Failed to dial WebSockets: unexpected EOF")
            {
                Error = "WEBSOCKETS: UNEXPECTED EOF";
            }
            else if (Error.Contains("Eine vorhandene Verbindung wurde vom Remotehost geschlossen."))
            {
                Error = "WEBSOCKETS: CONNECTION CLOSED";
            }

            #endregion
        }

        public static Result Parse(Dictionary<string, string> data)
        {
            return new Result
            {
                Id = int.Parse(data["id"]),
                Protocol = data["protocol"],
                Environment = data["enviroment"],
                TimeSlot = data["time_slot"],
                TestBegin = DateTime.Parse(data["test_begin"]),
                TestEnd = DateTime.Parse(data["test_end"]),
                ClientId = int.Parse(data["client_id"]),
                ParallelClients = int.Parse(data["parallel_clients"]),
                TransferStart = long.Parse(data["transfer_start_unix"]),
                TransferEnd = long.Parse(data["transfer_end_unix"]),
                BytesPayload = long.Parse(data["bytes_payload"]),
                CpuClientBefore = double.Parse(data["cpu_client_percent_before"], CultureInfo.InvariantCulture),
                CpuClientAfter = double.Parse(data["cpu_client_percent_after"], CultureInfo.InvariantCulture),
                CpuClientWhile = double.Parse(data["cpu_client_percent_while"], CultureInfo.InvariantCulture),
                CpuServerBefore = double.Parse(data["cpu_server_percent_before"], CultureInfo.InvariantCulture),
                CpuServerAfter = double.Parse(data["cpu_server_percent_after"], CultureInfo.InvariantCulture),
                CpuServerWhile = double.Parse(data["cpu_server_percent_while"], CultureInfo.InvariantCulture),
                RamClientBefore = long.Parse(data["ram_client_bytes_before"]),
                RamClientAfter = long.Parse(data["ram_client_bytes_after"]),
                RamClientWhile = long.Parse(data["ram_client_bytes_while"]),
                RamServerBefore = long.Parse(data["ram_server_bytes_before"]),
                RamServerAfter = long.Parse(data["ram_server_bytes_after"]),
                RamServerWhile = long.Parse(data["ram_server_bytes_while"]),
                LostPackets = long.Parse(data["lost_packets"]),
                Error = data["error"]
            };
        }
    }

    private static string GetResultsCsv()
    {
        return Path.Combine(Environment.CurrentDirectory, "..", "..", "..", "..", "assets", "results.csv");
    }
    
    private static string GetOutputCsv()
    {
        return Path.Combine(Environment.CurrentDirectory, "..", "..", "..", "..", "assets", "results_clean.csv");
    }
    
    private static long GetSecondsSinceEpoch(DateTime dateTime)
    {
        return (long)(dateTime - new DateTime(1970, 1, 1)).TotalSeconds;
    }
}