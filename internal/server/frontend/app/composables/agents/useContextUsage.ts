import { ref } from 'vue'

export interface ContextCategory {
  name: string
  tokens: number
  percentage: number
}

export interface ContextUsageData {
  model: string
  total_tokens: number
  context_window: number
  percentage: number
  categories: ContextCategory[]
}

export function useContextUsage() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  /**
   * Parses the /context command response to extract token usage by category
   * New format:
   * ## Context Usage
   * Model: claude-sonnet-4-5-20250929
   * Tokens: 47.6k / 200.0k (24%)
   *
   * ### Categories
   * | Category | Tokens | Percentage |
   * | System prompt | 2.6k | 1.3% |
   * | System tools | 12.4k | 6.2% |
   * | MCP tools | 1.7k | 0.9% |
   * | Memory files | 6.8k | 3.4% |
   * | Messages | 24.1k | 12.0% |
   * | Free space | 152.4k | 76.2% |
   */
  const parseContextResponse = (text: string): ContextUsageData | null => {
    try {
      // Remove XML tags if present (e.g., <local-command-stdout>)
      let cleanText = text.replace(/<[^>]+>/g, '')

      // Parse model
      const modelMatch = cleanText.match(/Model:\s*(.+)/i)
      if (!modelMatch) {
        return null
      }

      // Parse tokens: "74.2k / 200.0k (37%)"
      // Need to handle both with and without asterisks (markdown bold)
      const tokensMatch = cleanText.match(/Tokens:\*?\*?\s*([\d.]+)k\s*\/\s*([\d.]+)k\s*\((\d+)%\)/i)
      if (!tokensMatch) {
        return null
      }

      const parseTokenValue = (str: string): number => {
        const num = parseFloat(str)
        // Values are in thousands (e.g., "71.5k" = 71500)
        return num * 1000
      }

      const totalTokens = parseTokenValue(tokensMatch[1])
      const contextWindow = parseTokenValue(tokensMatch[2])
      const percentage = parseInt(tokensMatch[3], 10)

      // Parse categories table
      const categories: ContextCategory[] = []
      const categoryRegex = /\|\s*([^|]+?)\s*\|\s*([\d.]+)k\s*\|\s*([\d.]+)%\s*\|/gi

      let match
      while ((match = categoryRegex.exec(cleanText)) !== null) {
        const name = match[1].trim()
        const tokensStr = match[2].trim()
        const pct = parseFloat(match[3])

        // Skip header row and separator row
        if (name.toLowerCase() !== 'category' && !name.startsWith('---')) {
          categories.push({
            name,
            tokens: parseTokenValue(tokensStr),
            percentage: pct
          })
        }
      }

      if (categories.length === 0) {
        return null
      }

      return {
        model: modelMatch[1].trim(),
        total_tokens: totalTokens,
        context_window: contextWindow,
        percentage,
        categories
      }
    } catch (err) {
      console.error('Error parsing context response:', err)
      return null
    }
  }

  /**
   * Fetches context usage by sending a /context message to the agent
   * and waiting for the response
   */
  const fetchContextUsage = async (
    sessionId: string,
    sendMessage: (message: string) => Promise<void>,
    onResponse: (usage: ContextUsageData) => void
  ): Promise<void> => {
    if (loading.value) {
      return
    }

    loading.value = true
    error.value = null

    try {
      // Send /context command
      await sendMessage('/context')

      // Note: The response will be handled by the message handler
      // which should call parseContextResponse and then onResponse
    } catch (err) {
      console.error('Error fetching context usage:', err)
      error.value = err instanceof Error ? err.message : 'Failed to fetch context usage'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    parseContextResponse,
    fetchContextUsage,
  }
}
