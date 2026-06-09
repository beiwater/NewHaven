import { Toaster } from "@/components/ui/toaster"
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClientInstance } from '@/lib/query-client'
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import PageNotFound from './lib/PageNotFound';
import { AuthProvider, useAuth } from '@/lib/AuthContext';
import UserNotRegisteredError from '@/components/UserNotRegisteredError';
import ScrollToTop from './components/ScrollToTop';
import GameLayout from '@/components/game/GameLayout';
import MapPage from '@/pages/MapPage';
import BuildPage from '@/pages/BuildPage';
import MarketPage from '@/pages/MarketPage';
import OrdersPage from '@/pages/OrdersPage';
import ChatPage from '@/pages/ChatPage';
import WarehousePage from '@/pages/WarehousePage';
import FinancePage from '@/pages/FinancePage';
import ResearchPage from '@/pages/ResearchPage';
import ExecutivesPage from '@/pages/ExecutivesPage';
import LeaderboardPage from '@/pages/LeaderboardPage';
import AchievementsPage from '@/pages/AchievementsPage';
import SettingsPage from '@/pages/SettingsPage';
import ContractsPage from '@/pages/ContractsPage';
import CollectionPage from '@/pages/CollectionPage';
import WikiPage from '@/pages/WikiPage';
import MessagesPage from '@/pages/MessagesPage';
import PlayerProfilePage from '@/pages/PlayerProfilePage';

const AuthenticatedApp = () => {
  const { isLoadingAuth, isLoadingPublicSettings, authError, navigateToLogin } = useAuth();

  // Show loading spinner while checking app public settings or auth
  if (isLoadingPublicSettings || isLoadingAuth) {
    return (
      <div className="fixed inset-0 flex items-center justify-center">
        <div className="w-8 h-8 border-4 border-slate-200 border-t-slate-800 rounded-full animate-spin"></div>
      </div>
    );
  }

  // Handle authentication errors
  if (authError) {
    if (authError.type === 'user_not_registered') {
      return <UserNotRegisteredError />;
    } else if (authError.type === 'auth_required') {
      // Redirect to login automatically
      navigateToLogin();
      return null;
    }
  }

  // Render the main app
  return (
    <Routes>
      <Route element={<GameLayout />}>
        <Route path="/" element={<MapPage />} />
        <Route path="/build" element={<BuildPage />} />
        <Route path="/market" element={<MarketPage />} />
        <Route path="/orders" element={<OrdersPage />} />
        <Route path="/chat" element={<ChatPage />} />
        <Route path="/warehouse" element={<WarehousePage />} />
        <Route path="/finance" element={<FinancePage />} />
        <Route path="/research" element={<ResearchPage />} />
        <Route path="/executives" element={<ExecutivesPage />} />
        <Route path="/leaderboard" element={<LeaderboardPage />} />
        <Route path="/achievements" element={<AchievementsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/contracts" element={<ContractsPage />} />
        <Route path="/collection" element={<CollectionPage />} />
        <Route path="/wiki" element={<WikiPage />} />
        <Route path="/messages" element={<MessagesPage />} />
        <Route path="/profile" element={<PlayerProfilePage />} />
      </Route>
      <Route path="*" element={<PageNotFound />} />
    </Routes>
  );
};

function App() {

  return (
    <AuthProvider>
      <QueryClientProvider client={queryClientInstance}>
        <Router>
          <ScrollToTop />
          <AuthenticatedApp />
        </Router>
        <Toaster />
      </QueryClientProvider>
    </AuthProvider>
  )
}

export default App