import websocket


def on_message(ws, message):
    print(f"Received: {message}")
    ws.close()


def on_error(ws, error):
    print(f"Error: {error}")


def on_close(ws, close_status_code, close_msg):
    print("Closed")


def on_open(ws):
    print("Opened, requesting manifests...")
    ws.send('{"type":"request_manifests"}')


if __name__ == "__main__":
    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(
        "ws://localhost:8081/ws",
        on_open=on_open,
        on_message=on_message,
        on_error=on_error,
        on_close=on_close,
    )
    ws.run_forever()
