"use client";

import { useEffect, useRef, useState } from "react";

interface Message {
  id: string;
  sender: string;
  text: string;
  time: Date;
}

interface MessageStringDate {
  id: string;
  sender: string;
  text: string;
  time: string;
}

interface NewMessageEvent {
  type: "newMessage";
  data: MessageStringDate;
}

interface NewMessagesEvent {
  type: "newMessages";
  data: MessageStringDate[];
}

interface PingEvent {
  type: "ping";
}

interface OpenedEvent {
  type: "opened";
}

type WsEvent = NewMessageEvent | NewMessagesEvent | PingEvent | OpenedEvent;

const MessageRow = ({
  message,
  myUUID,
}: {
  message: Message;
  myUUID: string;
}) => {
  const mine = message.sender === myUUID;
  const seed = parseInt(message.sender.slice(0, 8), 16) / 0xffffffff;
  return (
    <div data-mine={mine} className="flex data-[mine=true]:justify-end">
      <div
        className="max-w-3xl m-2 p-4 rounded-lg"
        style={{ backgroundColor: `hsl(${~~(365 * seed)},80%,80%)` }}
      >
        {message.text}
      </div>
    </div>
  );
};

function tryClose(socket: WebSocket) {
  try {
    socket.close();
  } catch {}
}

export const RoomComponent = ({ roomId = "" }) => {
  const theirUUID = useRef(crypto.randomUUID());
  const nowRef = useRef(new Date());

  const myUUID = useRef(crypto.randomUUID());
  const [messages, setMessages] = useState<Message[]>([]);

  // Create WebSocket connection.
  const url = "ws://localhost:8081/ws";
  const socketRef = useRef<WebSocket>(null);
  const socketIdRef = useRef(0);

  const connectWs = () => {
    socketRef.current = new WebSocket(url);
    const socket = socketRef.current!;
    const myId = (socketIdRef.current += 1);

    // Listen for messages
    socket.addEventListener("message", (event) => {
      if (socketIdRef.current !== myId) return tryClose(socket);
      const decoded: WsEvent = JSON.parse(event.data);
      console.log("Message from server ", decoded);
    });

    const reconnect = () => {
      tryClose(socket);
      if (socketIdRef.current !== myId) return;
      setTimeout(() => {
        connectWs();
      }, 500);
    };
    socket.addEventListener("close", reconnect);
    socket.addEventListener("error", reconnect);
  };

  useEffect(() => {
    connectWs();
    setTimeout(() => {
      setMessages((old) => [
        ...old,
        {
          id: crypto.randomUUID(),
          sender: theirUUID.current!,
          text: "Hey what is uup",
          time: nowRef.current!,
        },
      ]);
    }, 1000);
  }, []);

  return (
    <div>
      {!!roomId && <div>Room: {roomId}</div>}
      <div>
        {messages.map((message) => (
          <MessageRow
            myUUID={myUUID.current!}
            message={message}
            key={message.id}
          />
        ))}
      </div>
    </div>
  );
};
