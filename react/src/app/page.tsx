"use client";

import { RoomComponent } from "@/components/room";
import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

export default function Home() {
  return (
    <Suspense>
      <Loaded />
    </Suspense>
  );
}
function Loaded() {
  const roomId = useSearchParams().get("roomId") || undefined;

  return <RoomComponent roomId={roomId} />;
}
