import { diffLines, type Change } from 'diff'

export interface DiffLine {
  type: 'addition' | 'deletion' | 'context'
  marker: string
  text: string
  lineNumber?: number
}

export interface DiffStats {
  additions: number
  deletions: number
  context: number
  total: number
}

/**
 * Composable for computing unified diffs with context
 */
export const useDiff = () => {
  /**
   * Compute a unified diff with context lines
   * @param oldText - Original text
   * @param newText - Modified text
   * @param contextLines - Number of context lines to show (default: 3)
   * @returns Array of diff lines with type, marker, and text
   */
  const computeDiff = (
    oldText: string,
    newText: string,
    contextLines: number = 3
  ): DiffLine[] => {
    if (!oldText && !newText) return []

    // Use jsdiff to compute changes
    const changes: Change[] = diffLines(oldText, newText)

    const result: DiffLine[] = []

    // Process each change
    for (const change of changes) {
      const lines = change.value.split('\n')

      // Remove last empty line if present (common artifact from split)
      if (lines.length > 0 && lines[lines.length - 1] === '') {
        lines.pop()
      }

      for (const line of lines) {
        if (change.added) {
          result.push({
            type: 'addition',
            marker: '+',
            text: line
          })
        } else if (change.removed) {
          result.push({
            type: 'deletion',
            marker: '-',
            text: line
          })
        } else {
          result.push({
            type: 'context',
            marker: ' ',
            text: line
          })
        }
      }
    }

    return result
  }

  /**
   * Compute diff with limited context (useful for large diffs)
   * Shows only N lines of context around changes
   */
  const computeContextualDiff = (
    oldText: string,
    newText: string,
    contextLines: number = 3
  ): DiffLine[] => {
    const allLines = computeDiff(oldText, newText)

    if (allLines.length <= 10) {
      // Small diff, show everything
      return allLines
    }

    // Find indices of changed lines
    const changedIndices = new Set<number>()
    allLines.forEach((line, idx) => {
      if (line.type !== 'context') {
        // Add this line and surrounding context
        for (let i = Math.max(0, idx - contextLines);
             i <= Math.min(allLines.length - 1, idx + contextLines);
             i++) {
          changedIndices.add(i)
        }
      }
    })

    // Build result with ellipsis for skipped sections
    const result: DiffLine[] = []
    let lastIncluded = -1

    Array.from(changedIndices).sort((a, b) => a - b).forEach(idx => {
      if (lastIncluded !== -1 && idx > lastIncluded + 1) {
        // Add ellipsis for skipped lines
        result.push({
          type: 'context',
          marker: '⋯',
          text: `... ${idx - lastIncluded - 1} lines hidden ...`
        })
      }
      result.push(allLines[idx])
      lastIncluded = idx
    })

    return result
  }

  /**
   * Calculate statistics about the diff
   */
  const getDiffStats = (diffLines: DiffLine[]): DiffStats => {
    const stats: DiffStats = {
      additions: 0,
      deletions: 0,
      context: 0,
      total: diffLines.length
    }

    for (const line of diffLines) {
      if (line.type === 'addition') {
        stats.additions++
      } else if (line.type === 'deletion') {
        stats.deletions++
      } else if (line.type === 'context') {
        stats.context++
      }
    }

    return stats
  }

  /**
   * Compute side-by-side diff (useful for split view)
   */
  const computeSideBySideDiff = (oldText: string, newText: string) => {
    const changes: Change[] = diffLines(oldText, newText)

    const leftLines: (DiffLine | null)[] = []
    const rightLines: (DiffLine | null)[] = []

    for (const change of changes) {
      const lines = change.value.split('\n').filter(l => l !== '')

      if (change.added) {
        // Only on right side
        lines.forEach(line => {
          leftLines.push(null)
          rightLines.push({
            type: 'addition',
            marker: '+',
            text: line
          })
        })
      } else if (change.removed) {
        // Only on left side
        lines.forEach(line => {
          leftLines.push({
            type: 'deletion',
            marker: '-',
            text: line
          })
          rightLines.push(null)
        })
      } else {
        // Context on both sides
        lines.forEach(line => {
          const contextLine = {
            type: 'context' as const,
            marker: ' ',
            text: line
          }
          leftLines.push(contextLine)
          rightLines.push(contextLine)
        })
      }
    }

    return { leftLines, rightLines }
  }

  return {
    computeDiff,
    computeContextualDiff,
    getDiffStats,
    computeSideBySideDiff
  }
}
