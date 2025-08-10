"use client";

const myUUID = crypto.randomUUID();
const theirUUID = crypto.randomUUID();
const now = new Date();

const messages = [
  {
    id: crypto.randomUUID(),
    sender: myUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 8 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: myUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 7 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: theirUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 6 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: theirUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 5 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: theirUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 4 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: myUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 3 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: theirUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 2 * 1000),
  },
  {
    id: crypto.randomUUID(),
    sender: myUUID,
    text: "Hey what is uup",
    time: new Date(now.getTime() - 1 * 1000),
  },
] as const;

const MessageRow = ({ message }: { message: (typeof messages)[number] }) => {
  return <div>{message.text}</div>;
};

export default function Home() {
  return (
    <div>
      {messages.map((message) => (
        <MessageRow message={message} key={message.id} />
      ))}
    </div>
  );
}
