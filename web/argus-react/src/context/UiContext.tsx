import React, { createContext, useContext, useReducer } from 'react';
import type { ReactNode } from 'react';
import type { ThemeMode } from '../types/common';
import type { ColorTone } from '../theme/theme';

// Define UI state types
interface UiState {
  darkMode: boolean;
  themeMode: ThemeMode;
  colorTone: ColorTone;
  sidebarOpen: boolean;
  refreshInterval: number;
  activeTab: string;
  notifications: Array<{
    id: string;
    message: string;
    type: 'info' | 'success' | 'warning' | 'error';
    timestamp: number;
  }>;
}

// Define action types
type UiAction =
  | { type: 'TOGGLE_DARK_MODE' }
  | { type: 'SET_DARK_MODE'; payload: boolean }
  | { type: 'SET_THEME_MODE'; payload: ThemeMode }
  | { type: 'SET_COLOR_TONE'; payload: ColorTone }
  | { type: 'TOGGLE_SIDEBAR' }
  | { type: 'SET_SIDEBAR'; payload: boolean }
  | { type: 'SET_REFRESH_INTERVAL'; payload: number }
  | { type: 'SET_ACTIVE_TAB'; payload: string }
  | { type: 'ADD_NOTIFICATION'; payload: { message: string; type: 'info' | 'success' | 'warning' | 'error' } }
  | { type: 'REMOVE_NOTIFICATION'; payload: string }
  | { type: 'CLEAR_NOTIFICATIONS' };

// Define the context interface
interface UiContextType {
  state: UiState;
  dispatch: React.Dispatch<UiAction>;
  toggleDarkMode: () => void;
  setThemeMode: (mode: ThemeMode) => void;
  setColorTone: (tone: ColorTone) => void;
  toggleSidebar: () => void;
  setRefreshInterval: (interval: number) => void;
  setActiveTab: (tab: string) => void;
  addNotification: (message: string, type: 'info' | 'success' | 'warning' | 'error') => void;
  removeNotification: (id: string) => void;
  clearNotifications: () => void;
}

// Create the context
const UiContext = createContext<UiContextType | undefined>(undefined);

// Initial state
const initialState: UiState = {
  darkMode: window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches,
  themeMode: (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) ? 'dark' : 'light',
  colorTone: 'morandi',
  sidebarOpen: true,
  refreshInterval: 5000, // 5 seconds
  activeTab: 'dashboard',
  notifications: [],
};

// Reducer function
function uiReducer(state: UiState, action: UiAction): UiState {
  switch (action.type) {
    case 'TOGGLE_DARK_MODE':
      return {
        ...state,
        darkMode: !state.darkMode,
        themeMode: !state.darkMode ? 'dark' : 'light',
      };
    case 'SET_DARK_MODE':
      return {
        ...state,
        darkMode: action.payload,
        themeMode: action.payload ? 'dark' : 'light',
      };
    case 'SET_THEME_MODE':
      return {
        ...state,
        themeMode: action.payload,
        darkMode: action.payload === 'dark',
      };
    case 'SET_COLOR_TONE':
      return {
        ...state,
        colorTone: action.payload,
      };
    case 'TOGGLE_SIDEBAR':
      return {
        ...state,
        sidebarOpen: !state.sidebarOpen,
      };
    case 'SET_SIDEBAR':
      return {
        ...state,
        sidebarOpen: action.payload,
      };
    case 'SET_REFRESH_INTERVAL':
      return {
        ...state,
        refreshInterval: action.payload,
      };
    case 'SET_ACTIVE_TAB':
      return {
        ...state,
        activeTab: action.payload,
      };
    case 'ADD_NOTIFICATION':
      return {
        ...state,
        notifications: [
          ...state.notifications,
          {
            id: Date.now().toString(),
            message: action.payload.message,
            type: action.payload.type,
            timestamp: Date.now(),
          },
        ],
      };
    case 'REMOVE_NOTIFICATION':
      return {
        ...state,
        notifications: state.notifications.filter(
          (notification) => notification.id !== action.payload
        ),
      };
    case 'CLEAR_NOTIFICATIONS':
      return {
        ...state,
        notifications: [],
      };
    default:
      return state;
  }
}

// Provider component
interface UiProviderProps {
  children: ReactNode;
  initialDarkMode?: boolean;
}

export const UiProvider: React.FC<UiProviderProps> = ({
  children,
  initialDarkMode,
}) => {
  const [state, dispatch] = useReducer(uiReducer, {
    ...initialState,
    darkMode: initialDarkMode !== undefined ? initialDarkMode : initialState.darkMode,
    themeMode: initialDarkMode !== undefined ? (initialDarkMode ? 'dark' : 'light') : initialState.themeMode,
  });

  // Helper functions
  const toggleDarkMode = () => dispatch({ type: 'TOGGLE_DARK_MODE' });
  const setThemeMode = (mode: ThemeMode) => dispatch({ type: 'SET_THEME_MODE', payload: mode });
  const setColorTone = (tone: ColorTone) => dispatch({ type: 'SET_COLOR_TONE', payload: tone });
  const toggleSidebar = () => dispatch({ type: 'TOGGLE_SIDEBAR' });
  const setRefreshInterval = (interval: number) => dispatch({ type: 'SET_REFRESH_INTERVAL', payload: interval });
  const setActiveTab = (tab: string) => dispatch({ type: 'SET_ACTIVE_TAB', payload: tab });
  
  const addNotification = (message: string, type: 'info' | 'success' | 'warning' | 'error') => 
    dispatch({ type: 'ADD_NOTIFICATION', payload: { message, type } });
  
  const removeNotification = (id: string) => dispatch({ type: 'REMOVE_NOTIFICATION', payload: id });
  const clearNotifications = () => dispatch({ type: 'CLEAR_NOTIFICATIONS' });

  return (
    <UiContext.Provider
      value={{
        state,
        dispatch,
        toggleDarkMode,
        setThemeMode,
        setColorTone,
        toggleSidebar,
        setRefreshInterval,
        setActiveTab,
        addNotification,
        removeNotification,
        clearNotifications,
      }}
    >
      {children}
    </UiContext.Provider>
  );
};

// Custom hook for using the UI context
export const useUiContext = () => {
  const context = useContext(UiContext);
  
  if (context === undefined) {
    throw new Error('useUiContext must be used within a UiProvider');
  }
  
  return context;
};

export default UiContext; 