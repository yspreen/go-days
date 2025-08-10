"use client";

import { useEffect, useRef, useState } from "react";

import { Input } from "./ui/input";

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

interface AuthResponseEvent {
  type: "authResponse";
  secret: string;
  userId: string;
}

type WsEvent =
  | NewMessageEvent
  | NewMessagesEvent
  | PingEvent
  | OpenedEvent
  | AuthResponseEvent;

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

function parseWsMessage(msg: MessageStringDate) {
  return {
    ...msg,
    time: new Date(msg.time),
  };
}

export const RoomComponent = ({ roomId = "" }) => {
  const [myUuid, setMyUuid] = useState("");
  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");

  // Scroll management for the messages list
  const listRef = useRef<HTMLDivElement>(null);
  const atBottomRef = useRef(true);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const SCROLL_THRESHOLD = 24; // px tolerance

  const updateIsAtBottom = () => {
    const el = listRef.current;
    if (!el) return;
    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    const atBottom = distanceFromBottom <= SCROLL_THRESHOLD;
    atBottomRef.current = atBottom;
    setIsAtBottom(atBottom);
  };

  const scrollToBottom = (smooth = false) => {
    const el = listRef.current;
    if (!el) return;
    const behavior: ScrollBehavior = smooth ? "smooth" : "auto";
    // Ensure DOM is updated before scrolling
    requestAnimationFrame(() => {
      el.scrollTo({ top: el.scrollHeight, behavior });
    });
  };

  // Create WebSocket connection.
  const host =
    process.env.NODE_ENV === "production"
      ? "wss://go-days.fly.dev"
      : "ws://localhost:8081";
  const url = `${host}/ws`;
  const socketRef = useRef<WebSocket>(null);
  const socketIdRef = useRef(0);

  const handleMessage = (event: WsEvent) => {
    if (event.type === "authResponse") {
      setMyUuid(event.userId);
      localStorage.setItem("secret", event.secret);
    }
    if (event.type === "newMessage") {
      setMessages((old) => [...old, parseWsMessage(event.data)]);
    }
    if (event.type === "newMessages") {
      setMessages((old) => [...old, ...event.data.map(parseWsMessage)]);
    }
  };

  const connectWs = () => {
    socketRef.current = new WebSocket(url);
    const socket = socketRef.current!;
    const myId = (socketIdRef.current += 1);

    // Listen for messages
    socket.addEventListener("message", (event) => {
      if (socketIdRef.current !== myId) return tryClose(socket);

      handleMessage(JSON.parse(event.data));
    });

    socket.addEventListener("open", () => {
      setMessages([]);
      // Reset scroll state on new connection
      atBottomRef.current = true;
      setIsAtBottom(true);
      socket.send(
        JSON.stringify({
          type: "auth",
          secret: localStorage.getItem("secret") || "",
          roomId,
        })
      );
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
  }, []);

  // Attach scroll listener to track when user is at the bottom
  useEffect(() => {
    const el = listRef.current;
    if (!el) return;
    updateIsAtBottom();
    const onScroll = () => updateIsAtBottom();
    el.addEventListener("scroll", onScroll);
    return () => {
      el.removeEventListener("scroll", onScroll);
    };
  }, []);

  // Auto-scroll on new messages only if currently at bottom
  useEffect(() => {
    if (!atBottomRef.current) return;
    scrollToBottom(false);
  }, [messages]);

  const send = () => {
    socketRef.current?.send(
      JSON.stringify({
        type: "send",
        message,
      })
    );
    setMessage("");
  };

  return (
    <div className="flex flex-col h-svh max-w-3xl mx-auto">
      {!!roomId && <div>Room: {roomId}</div>}
      <div ref={listRef} className="relative flex-1 overflow-y-auto">
        {messages.map((message) => (
          <MessageRow myUUID={myUuid} message={message} key={message.id} />
        ))}
      </div>
      <div className="p-2">
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            send();
          }}
        >
          <div className="*:not-first:mt-2">
            <div className="flex rounded-md shadow-xs">
              <Input
                className="-me-px flex-1 rounded-e-none shadow-none focus-visible:z-10"
                placeholder="Message"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
              />
              <button className="border-input bg-background text-foreground hover:bg-accent hover:text-foreground focus-visible:border-ring focus-visible:ring-ring/50 inline-flex items-center rounded-e-md border px-3 text-sm font-medium transition-[color,box-shadow] outline-none focus:z-10 focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50">
                Send
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};
