// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import { ExtensionContext } from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: ExtensionContext) {
    // The server is implemented in node
    // const serverModule = context.asAbsolutePath(
    //     path.join('bin', 'protobuf-language-server')
    // );
    const serverModule = 'protobuf-language-server';
    // let debugOptions = { execArgv: ["--nolazy", "--debug=6004"] };
    // If the extension is launched in debug mode then the debug server options are used
    // Otherwise the run options are used
    const serverOptions: ServerOptions = {
        run: { command: serverModule, transport: TransportKind.stdio },
        debug: { command: serverModule, transport: TransportKind.stdio },
    };

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        // Register the server for plain text documents
        documentSelector: [{ scheme: 'file', language: 'proto' }],
    };

    // Create the language client and start the client.
    let client = new LanguageClient(
        'proto',
        'protobuf-language-server',
        serverOptions,
        clientOptions
    );

    // Start the client. This will also launch the server
    client.start();
}

// this method is called when your extension is deactivated
export function deactivate() { }