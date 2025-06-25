/**
 * Design Token Service for Figma Integration
 * Manages dynamic styling from Figma design tokens
 */

import { designSystemApi } from './api/designSystemApi';

export interface DesignToken {
  name: string;
  value: string | number | object;
  type: 'color' | 'spacing' | 'typography' | 'shadow' | 'border' | 'size';
  category: string;
  description?: string;
  figmaId?: string;
}

export interface FigmaColorToken extends DesignToken {
  type: 'color';
  value: string; // hex, rgb, hsl
  opacity?: number;
}

export interface FigmaSpacingToken extends DesignToken {
  type: 'spacing';
  value: number; // pixels
  unit: 'px' | 'rem' | 'em';
}

export interface FigmaTypographyToken extends DesignToken {
  type: 'typography';
  value: {
    fontFamily: string;
    fontSize: number;
    fontWeight: number;
    lineHeight: number;
    letterSpacing?: number;
  };
}

export interface FigmaShadowToken extends DesignToken {
  type: 'shadow';
  value: {
    x: number;
    y: number;
    blur: number;
    spread: number;
    color: string;
    inset?: boolean;
  }[];
}

export interface StyleTheme {
  id: string;
  name: string;
  description: string;
  tokens: DesignToken[];
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
  figmaFileId?: string;
  figmaNodeId?: string;
}

interface FigmaColorData {
  value?: string;
  r?: number;
  g?: number;
  b?: number;
  id?: string;
  description?: string;
}

interface FigmaSpacingData {
  value: number;
  id?: string;
  description?: string;
}

interface FigmaTypographyData {
  fontFamily: string;
  fontSize: number;
  fontWeight: number;
  lineHeight: number;
  letterSpacing?: number;
  id?: string;
  description?: string;
}

interface FigmaShadowData {
  shadows?: any[];
  value?: any;
  id?: string;
  description?: string;
}

class DesignTokenService {
  private currentTheme: StyleTheme | null = null;
  private themes: StyleTheme[] = [];
  private cssRoot: HTMLElement | null = null;

  constructor() {
    this.cssRoot = document.documentElement;
    this.loadThemes();
  }

  /**
   * Parse Figma design tokens from API response
   */
  async parseFigmaTokens(figmaData: any): Promise<DesignToken[]> {
    // This method is now handled by the backend API
    // Keep for compatibility but delegate to backend
    return [];
  }

  /**
   * Apply design tokens to CSS custom properties
   */
  applyTokensToCSS(tokens: DesignToken[]): void {
    if (!this.cssRoot) return;

    tokens.forEach(token => {
      const cssProperty = this.tokenToCSSProperty(token);
      const cssValue = this.tokenToCSSValue(token);
      
      this.cssRoot!.style.setProperty(cssProperty, cssValue);
    });

    // Trigger custom event for components to react to theme changes
    window.dispatchEvent(new CustomEvent('themeChanged', {
      detail: { tokens }
    }));
  }

  /**
   * Create new theme from Figma data
   */
  async createThemeFromFigma(figmaFileId: string, figmaNodeId?: string): Promise<StyleTheme> {
    try {
      const theme = await designSystemApi.createThemeFromFigma({
        figmaFileId,
        figmaNodeId,
      });

      // Update local cache
      this.themes.push(theme);
      this.saveThemesToCache();

      return theme;
    } catch (error) {
      console.error('Failed to create theme from Figma:', error);
      throw error;
    }
  }

  /**
   * Activate a theme
   */
  async activateTheme(themeId: string): Promise<void> {
    try {
      const theme = await designSystemApi.activateTheme(parseInt(themeId));
      
      // Update local state
      this.themes.forEach(t => t.isActive = false);
      const localTheme = this.themes.find(t => t.id.toString() === themeId);
      if (localTheme) {
        localTheme.isActive = true;
        this.currentTheme = localTheme;
      } else {
        // Add theme to local cache if not present
        this.themes.push(theme);
        this.currentTheme = theme;
      }

      // Apply tokens to CSS
      this.applyTokensToCSS(theme.tokens);
      
      // Save state
      this.saveThemesToCache();
      localStorage.setItem('activeThemeId', themeId);
    } catch (error) {
      console.error('Failed to activate theme:', error);
      throw error;
    }
  }

  /**
   * Update theme with new Figma data
   */
  async updateThemeFromFigma(themeId: string): Promise<StyleTheme> {
    try {
      const theme = await designSystemApi.syncThemeWithFigma(parseInt(themeId));

      // Update local cache
      const index = this.themes.findIndex(t => t.id.toString() === themeId);
      if (index !== -1) {
        this.themes[index] = theme;
        if (theme.isActive) {
          this.currentTheme = theme;
          this.applyTokensToCSS(theme.tokens);
        }
      }

      this.saveThemesToCache();
      return theme;
    } catch (error) {
      console.error('Failed to update theme from Figma:', error);
      throw error;
    }
  }

  /**
   * Get all available themes
   */
  async getThemes(): Promise<StyleTheme[]> {
    try {
      this.themes = await designSystemApi.getThemes();
      this.saveThemesToCache();
      return this.themes;
    } catch (error) {
      console.error('Failed to fetch themes:', error);
      // Return cached themes as fallback
      return this.themes;
    }
  }

  /**
   * Get current active theme
   */
  async getCurrentTheme(): Promise<StyleTheme | null> {
    try {
      this.currentTheme = await designSystemApi.getActiveTheme();
      if (this.currentTheme) {
        this.applyTokensToCSSSilently(this.currentTheme.tokens);
      }
      return this.currentTheme;
    } catch (error) {
      console.error('Failed to fetch current theme:', error);
      return this.currentTheme;
    }
  }

  /**
   * Delete theme
   */
  async deleteTheme(themeId: string): Promise<void> {
    try {
      await designSystemApi.deleteTheme(parseInt(themeId));
      
      // Update local cache
      const index = this.themes.findIndex(t => t.id.toString() === themeId);
      if (index !== -1) {
        const theme = this.themes[index];
        if (theme.isActive) {
          this.revertToDefaultTheme();
          this.currentTheme = null;
        }
        this.themes.splice(index, 1);
      }

      this.saveThemesToCache();
    } catch (error) {
      console.error('Failed to delete theme:', error);
      throw error;
    }
  }

  /**
   * Export theme configuration
   */
  async exportTheme(themeId: string): Promise<string> {
    try {
      const blob = await designSystemApi.exportTheme(parseInt(themeId));
      return await blob.text();
    } catch (error) {
      console.error('Failed to export theme:', error);
      throw error;
    }
  }

  /**
   * Import theme configuration
   */
  async importTheme(themeData: string): Promise<StyleTheme> {
    try {
      const theme = await designSystemApi.importTheme(themeData);
      
      // Update local cache
      this.themes.push(theme);
      this.saveThemesToCache();

      return theme;
    } catch (error) {
      console.error('Failed to import theme:', error);
      throw error;
    }
  }

  // Local methods that don't require API calls

  /**
   * Get themes from local cache (synchronous)
   */
  getThemesSync(): StyleTheme[] {
    return this.themes;
  }

  /**
   * Get current theme from local cache (synchronous)
   */
  getCurrentThemeSync(): StyleTheme | null {
    return this.currentTheme;
  }

  // Private methods

  private async loadThemes(): Promise<void> {
    // Load from cache first for immediate UI
    this.loadThemesFromCache();
    
    // Then sync with backend
    try {
      await this.getThemes();
      await this.getCurrentTheme();
    } catch (error) {
      console.warn('Failed to sync themes with backend, using cached data');
    }
  }

  private loadThemesFromCache(): void {
    const stored = localStorage.getItem('designThemes');
    if (stored) {
      try {
        this.themes = JSON.parse(stored);
        
        // Load active theme from cache
        const activeThemeId = localStorage.getItem('activeThemeId');
        if (activeThemeId) {
          const activeTheme = this.themes.find(t => t.id.toString() === activeThemeId);
          if (activeTheme) {
            this.currentTheme = activeTheme;
            this.applyTokensToCSSSilently(activeTheme.tokens);
          }
        }
      } catch (error) {
        console.error('Failed to load themes from cache:', error);
        this.themes = [];
      }
    }
  }

  private saveThemesToCache(): void {
    localStorage.setItem('designThemes', JSON.stringify(this.themes));
  }

  private applyTokensToCSSSilently(tokens: DesignToken[]): void {
    if (!this.cssRoot) return;

    tokens.forEach(token => {
      const cssProperty = this.tokenToCSSProperty(token);
      const cssValue = this.tokenToCSSValue(token);
      
      this.cssRoot!.style.setProperty(cssProperty, cssValue);
    });
    
    // Don't trigger event for silent application
  }

  private tokenToCSSProperty(token: DesignToken): string {
    const prefix = '--';
    return `${prefix}${token.name}`;
  }

  private tokenToCSSValue(token: DesignToken): string {
    switch (token.type) {
      case 'color':
        return token.value as string;
      case 'spacing':
        const spacingToken = token as FigmaSpacingToken;
        return `${spacingToken.value}${spacingToken.unit || 'px'}`;
      case 'typography':
        const typographyToken = token as FigmaTypographyToken;
        return JSON.stringify(typographyToken.value);
      case 'shadow':
        const shadowToken = token as FigmaShadowToken;
        return this.shadowArrayToCSSString(shadowToken.value);
      default:
        return String(token.value);
    }
  }

  private shadowArrayToCSSString(shadows: any[]): string {
    return shadows.map(shadow => {
      const inset = shadow.inset ? 'inset ' : '';
      return `${inset}${shadow.x}px ${shadow.y}px ${shadow.blur}px ${shadow.spread}px ${shadow.color}`;
    }).join(', ');
  }

  private revertToDefaultTheme(): void {
    // Apply default CSS custom properties
    if (this.cssRoot) {
      // Remove all custom theme properties and let CSS defaults take over
      const computedStyle = getComputedStyle(this.cssRoot);
      const customProps = Array.from(computedStyle).filter(prop => prop.startsWith('--'));
      
      customProps.forEach(prop => {
        this.cssRoot!.style.removeProperty(prop);
      });
    }
  }
}

export const designTokenService = new DesignTokenService(); 