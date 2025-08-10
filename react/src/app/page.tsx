"use client";

import { useEffect, useRef, useState } from "react";

interface Message {
  id: string;
  sender: string;
  text: string;
  time: Date;
}

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

export default function Home() {
  const theirUUID = useRef(crypto.randomUUID());
  const nowRef = useRef(new Date());

  const myUUID = useRef(crypto.randomUUID());
  const [messages, setMessages] = useState<Message[]>([]);

  useEffect(() => {
    setMessages([
      {
        id: crypto.randomUUID(),
        sender: myUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 8 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: myUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 7 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: theirUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 6 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: theirUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 5 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: theirUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 4 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: myUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 3 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: theirUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 2 * 1000),
      },
      {
        id: crypto.randomUUID(),
        sender: myUUID.current!,
        text: "Hey what is uup",
        time: new Date(nowRef.current!.getTime() - 1 * 1000),
      },
    ]);
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
      {messages.map((message) => (
        <MessageRow
          myUUID={myUUID.current!}
          message={message}
          key={message.id}
        />
      ))}
    </div>
  );
}
