/**
 * Composable for mapping session names to South Park character avatars
 */

export interface CharacterInfo {
  name: string
  avatar: string
  color: string
}

// Character mapping (case-insensitive)
const characterMap: Record<string, CharacterInfo> = {
  stan: {
    name: 'Stan Marsh',
    avatar: '/avatars/stan.png',
    color: '#4A90E2'
  },
  kyle: {
    name: 'Kyle Broflovski',
    avatar: '/avatars/kyle.png',
    color: '#27AE60'
  },
  cartman: {
    name: 'Eric Cartman',
    avatar: '/avatars/cartman.png',
    color: '#E74C3C'
  },
  kenny: {
    name: 'Kenny McCormick',
    avatar: '/avatars/kenny.png',
    color: '#F39C12'
  },
  butters: {
    name: 'Butters Stotch',
    avatar: '/avatars/butters.png',
    color: '#F9D71C'
  },
  ike: {
    name: 'Ike Broflovski',
    avatar: '/avatars/ike.png',
    color: '#9B59B6'
  },
  lke: {
    name: 'Ike Broflovski',
    avatar: '/avatars/ike.png',
    color: '#9B59B6'
  },
  token: {
    name: 'Token Black',
    avatar: '/avatars/token.png',
    color: '#34495E'
  },
  wendy: {
    name: 'Wendy Testaburger',
    avatar: '/avatars/wendy.png',
    color: '#E91E63'
  },
  timmy: {
    name: 'Timmy Burch',
    avatar: '/avatars/timmy.png',
    color: '#3498DB'
  },
  jimmy: {
    name: 'Jimmy Valmer',
    avatar: '/avatars/jimmy.png',
    color: '#16A085'
  },
  randy: {
    name: 'Randy Marsh',
    avatar: '/avatars/randy.png',
    color: '#8E44AD'
  },
  tweek: {
    name: 'Tweek Tweak',
    avatar: '/avatars/Tweek.png',
    color: '#D4A017'
  },
  craig: {
    name: 'Craig Tucker',
    avatar: '/avatars/Craig.png',
    color: '#2C3E50'
  },
  sheila: {
    name: 'Sheila Broflovski',
    avatar: '/avatars/Sheila.png',
    color: '#E67E22'
  },
  sharon: {
    name: 'Sharon Marsh',
    avatar: '/avatars/Sharon.png',
    color: '#C0392B'
  },
  chef: {
    name: 'Chef',
    avatar: '/avatars/Chef.png',
    color: '#8B4513'
  },
  'mr-garrison': {
    name: 'Mr. Garrison',
    avatar: '/avatars/Mr-Garrison.png',
    color: '#7F8C8D'
  },
  'mr-mackey': {
    name: 'Mr. Mackey',
    avatar: '/avatars/Mr-Mackey.png',
    color: '#A0522D'
  },
  mackey: {
    name: 'Mr. Mackey',
    avatar: '/avatars/Mr-Mackey.png',
    color: '#A0522D'
  },
  bebe: {
    name: 'Bebe Stevens',
    avatar: '/avatars/Bebe.png',
    color: '#FF69B4'
  },
  clyde: {
    name: 'Clyde Donovan',
    avatar: '/avatars/Clyde.png',
    color: '#5DADE2'
  },
  'pc-principal': {
    name: 'PC Principal',
    avatar: '/avatars/PC-Principal.png',
    color: '#1ABC9C'
  },
  towelie: {
    name: 'Towelie',
    avatar: '/avatars/Towelie.png',
    color: '#34495E'
  },
  'mr-hankey': {
    name: 'Mr. Hankey',
    avatar: '/avatars/Mr-Hankey.png',
    color: '#8B4513'
  },
  'big-gay-al': {
    name: 'Big Gay Al',
    avatar: '/avatars/Big-Gay-Al.png',
    color: '#E91E63'
  },
  satan: {
    name: 'Satan',
    avatar: '/avatars/Satan.png',
    color: '#C0392B'
  }
}

const defaultCharacter: CharacterInfo = {
  name: 'Unknown',
  avatar: '/avatars/default.png',
  color: '#95A5A6'
}

/**
 * Get character avatar information for a session name
 * @param sessionName - The session name to look up (case-insensitive)
 * @returns Character information including name, avatar path, and color
 */
export function useCharacterAvatar(sessionName?: string | null): CharacterInfo {
  if (!sessionName) {
    return defaultCharacter
  }

  // Normalize session name (lowercase, trim whitespace)
  const normalizedName = sessionName.toLowerCase().trim()

  // Return character info or default
  return characterMap[normalizedName] || defaultCharacter
}

/**
 * Get all available characters
 * @returns Array of all character information
 */
export function getAllCharacters(): CharacterInfo[] {
  return Object.values(characterMap).filter(
    (char, index, self) => self.findIndex(c => c.name === char.name) === index
  )
}

/**
 * Check if a session name has a mapped character
 * @param sessionName - The session name to check
 * @returns True if character exists, false otherwise
 */
export function hasCharacter(sessionName?: string | null): boolean {
  if (!sessionName) return false
  const normalizedName = sessionName.toLowerCase().trim()
  return normalizedName in characterMap
}
