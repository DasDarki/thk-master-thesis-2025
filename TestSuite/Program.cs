using Spectre.Console;

namespace TestSuite;

internal class Program
{
    private static void Main(string[] args)
    {
        var local = AnsiConsole.Prompt(new SelectionPrompt<string>()
            .Title("Which environment do you want to run the tests in?")
            .AddChoices(Tester.EnvironmentLocal, Tester.EnvironmentRemote)) == Tester.EnvironmentLocal;
        
        var timeSlot = AnsiConsole.Prompt(new SelectionPrompt<string>()
            .Title("Which time slot do you want to run the tests in?")
            .AddChoices(Tester.TimeSlotMorning, Tester.TimeSlotAfternoon, Tester.TimeSlotEvening, Tester.TimeSlotNight));

        var protocols = AnsiConsole.Prompt(new MultiSelectionPrompt<string>()
            .Title("Which protocol(s) do you want to test?")
            .AddChoices(Tester.ProtocolHTTP3, Tester.ProtocolWebTransport, Tester.ProtocolWebSockets,
                Tester.ProtocolWebRTC));

        var parallelClients = AnsiConsole.Prompt(new MultiSelectionPrompt<int>()
            .Title("How many parallel client runs do you want to run?")
            .AddChoices(1, 5, 10, 20));
        
        Console.Clear();

        try
        {
            foreach (var protocol in protocols)
            {
                Console.WriteLine();
                Console.WriteLine();
                AnsiConsole.Write(new Markup("[bold blue]Running tests for protocol: " + protocol + "[/]\n"));

                foreach (var parallelClient in parallelClients)
                {
                    AnsiConsole.Write(new Markup("[bold green]Running " + parallelClient +
                                                 " parallel clients...[/]\n"));

                    var client = GetTestClient(protocol);
                    Tester.Run(client, local, timeSlot, parallelClient);

                    AnsiConsole.Write(new Markup("[bold green]Finished running " + parallelClient +
                                                 " parallel clients.[/]\n"));
                }

                AnsiConsole.Write(new Markup("[bold green]Finished running tests for protocol: " + protocol + "[/]\n"));
            }
        }
        finally
        {
            Console.WriteLine();
            Console.WriteLine();
            AnsiConsole.Write(new Markup("[bold yellow]Press enter key to exit...[/]\n"));
            Console.ReadLine();
        }
    }
    
    private static TestClient GetTestClient(string protocol)
    {
        return protocol switch
        {
            Tester.ProtocolHTTP3 => Tester.Http3Client,
            Tester.ProtocolWebTransport => Tester.WebTransportClient,
            Tester.ProtocolWebSockets => Tester.WebSocketsClient,
            Tester.ProtocolWebRTC => Tester.WebRTCClient,
            _ => throw new ArgumentException("Invalid protocol: " + protocol)
        };
    }
}