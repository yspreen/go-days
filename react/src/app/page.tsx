"use client";

import { RoomComponent } from "@/components/room";
import { useSearchParams } from "next/navigation";

export default function Home() {
  const roomId = useSearchParams().get("roomId") || undefined;

  return <RoomComponent roomId={roomId} />;
}
