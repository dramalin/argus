import React, { ReactNode } from 'react';
import { MetricsProvider } from './MetricsContext';
import { ProcessesProvider } from './ProcessesContext';
import { UiProvider } from './UiContext';

interface AppProviderProps {
  children: ReactNode;
}

/**
 * AppProvider combines all context providers in the application
 * This ensures proper nesting and allows components to access any context
 */
export const AppProvider: React.FC<AppProviderProps> = ({ children }) => {
  return (
    <UiProvider>
      <MetricsProvider>
        <ProcessesProvider>
          {children}
        </ProcessesProvider>
      </MetricsProvider>
    </UiProvider>
  );
};

export default AppProvider; 