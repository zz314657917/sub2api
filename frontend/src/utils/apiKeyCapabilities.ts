import type { ApiKey, Group } from '@/types'

type RoutePredicate = (route: ApiKey['multi_group_routes'][number]) => boolean

function apiKeyGroupsForRoutes(key: ApiKey, routePredicate: RoutePredicate = () => true): Group[] {
  const groups: Group[] = []
  const seen = new Set<number>()
  const append = (group: Group | undefined | null) => {
    if (!group || group.status === 'inactive' || seen.has(group.id)) return
    seen.add(group.id)
    groups.push(group)
  }

  append(key.group)
  const routeGroupIds = new Set<number>()
  for (const route of key.multi_group_routes || []) {
    if (!route.enabled || !routePredicate(route)) continue
    routeGroupIds.add(route.group_id)
  }
  if (routeGroupIds.size > 0 && Array.isArray(key.route_groups)) {
    key.route_groups.forEach((group) => {
      if (routeGroupIds.has(group.id)) append(group)
    })
  }

  return groups
}

export function apiKeyGroups(key: ApiKey): Group[] {
  return apiKeyGroupsForRoutes(key)
}

export function apiKeyChatGroups(key: ApiKey): Group[] {
  return apiKeyGroupsForRoutes(key, (route) => route.image_only !== true).filter((group) =>
    group.platform === 'openai' ||
    group.platform === 'anthropic' ||
    group.platform === 'gemini' ||
    group.platform === 'antigravity'
  )
}

export function apiKeySupportsChat(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyChatGroups(key).length > 0
}

export function apiKeySupportsOpenAI(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyGroups(key).some((group) => group.platform === 'openai')
}

export function apiKeyOpenAIImageGroups(key: ApiKey): Group[] {
  return apiKeyGroupsForRoutes(key, (route) => route.text_only !== true).filter((group) =>
    group.platform === 'openai' && group.allow_image_generation === true
  )
}

export function apiKeySupportsOpenAIImageGeneration(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyOpenAIImageGroups(key).length > 0
}

export function primaryAPIKeyGroupName(key: ApiKey): string {
  return key.group?.name || apiKeyGroups(key)[0]?.name || ''
}

export function primaryAPIKeyImageGroupName(key: ApiKey): string {
  return apiKeyOpenAIImageGroups(key)[0]?.name || primaryAPIKeyGroupName(key)
}
