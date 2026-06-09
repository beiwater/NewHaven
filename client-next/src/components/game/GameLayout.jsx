import { Outlet, useLocation } from "react-router-dom";
import TopBar from "./TopBar";
import Sidebar from "./Sidebar";
import PriceTicker from "./PriceTicker";
import MobileSecondaryNav from "./MobileSecondaryNav";

export default function GameLayout() {
  return (
    <div className="min-h-screen bg-background">
      <TopBar />
      <Sidebar />
      <main className="pt-14 md:pl-[200px] pb-12 md:pb-10 min-h-screen">
        <MobileSecondaryNav />
        <Outlet />
      </main>
      <PriceTicker />
    </div>
  );
}