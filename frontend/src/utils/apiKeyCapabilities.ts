import type { ApiKey, Group } from '@/types'

type RoutePredicate = (route: ApiKey['multi_group_routes'][number]) => boolean

function groupSupportsChat(group: Group | undefined | null): group is Group {
  if (!group || group.status === 'inactive') return false
  return group.routing_scope === 'inference' && (
    group.platform === 'openai' ||
    group.platform === 'anthropic' ||
    group.platform === 'gemini' ||
    group.platform === 'antigravity'
  )
}

function groupSupportsOpenAIImage(group: Group | undefined | null): group is Group {
  return !!group &&
    group.status !== 'inactive' &&
    group.platform === 'openai' &&
    group.routing_scope === 'image' &&
    group.allow_image_generation === true
}

function groupSupportsVideo(group: Group | undefined | null): group is Group {
  return !!group && group.status !== 'inactive' && group.routing_scope === 'video'
}

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
  return apiKeyGroupsForRoutes(key, (route) => route.image_only !== true).filter(groupSupportsChat)
}

export function apiKeySupportsChat(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyChatGroups(key).length > 0
}

export function apiKeySupportsOpenAI(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyGroups(key).some((group) => group.platform === 'openai')
}

export function apiKeyOpenAIImageGroups(key: ApiKey): Group[] {
  return apiKeyGroupsForRoutes(key, (route) => route.text_only !== true).filter(groupSupportsOpenAIImage)
}

export function apiKeySupportsOpenAIImageGeneration(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyOpenAIImageGroups(key).length > 0
}

function routeLooksVideoCapable(route: ApiKey['multi_group_routes'][number], group?: Group | null): boolean {
  if (!route.enabled || route.image_only === true || route.text_only === true) return false
  return groupSupportsVideo(group)
}

export function apiKeyVideoGroups(key: ApiKey): Group[] {
  const groups: Group[] = []
  const seen = new Set<number>()
  const groupsByID = new Map<number, Group>()
  if (key.group) groupsByID.set(key.group.id, key.group)
  for (const group of key.route_groups || []) {
    groupsByID.set(group.id, group)
  }
  if (groupSupportsVideo(key.group)) {
    seen.add(key.group.id)
    groups.push(key.group)
  }
  for (const route of key.multi_group_routes || []) {
    const group = groupsByID.get(route.group_id)
    if (!group || group.status === 'inactive' || seen.has(group.id)) continue
    if (!routeLooksVideoCapable(route, group)) continue
    seen.add(group.id)
    groups.push(group)
  }
  return groups
}

export function apiKeySupportsVideoGeneration(key: ApiKey): boolean {
  return key.status === 'active' && apiKeyVideoGroups(key).length > 0
}

export function apiKeySupportsUnifiedAccess(key: ApiKey): boolean {
  return apiKeySupportsChat(key) &&
    apiKeySupportsOpenAIImageGeneration(key) &&
    apiKeySupportsVideoGeneration(key)
}

export function primaryAPIKeyGroupName(key: ApiKey): string {
  return key.group?.name || apiKeyGroups(key)[0]?.name || ''
}

export function primaryAPIKeyImageGroupName(key: ApiKey): string {
  return apiKeyOpenAIImageGroups(key)[0]?.name || primaryAPIKeyGroupName(key)
}
