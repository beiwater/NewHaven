// Redirect to chat with DM tab
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
export default function MessagesPage() {
  const nav = useNavigate();
  useEffect(() => { nav("/chat"); }, []);
  return null;
}