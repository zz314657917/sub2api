<template>
  <div v-if="rows.length > 0" class="system-metadata-details mt-3">
    <div class="mb-2 text-xs font-semibold uppercase tracking-wide opacity-70">
      {{ t('tickets.metadata.changeDetails') }}
    </div>
    <dl class="grid gap-2">
      <div v-for="row in rows" :key="row.key" class="metadata-row">
        <dt class="metadata-label">{{ row.label }}</dt>
        <dd class="metadata-value">
          <template v-if="row.oldValue">
            <span class="metadata-old">{{ row.oldValue }}</span>
            <span class="metadata-arrow">→</span>
          </template>
          <span>{{ row.newValue }}</span>
        </dd>
      </div>
    </dl>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SupportTicketMessage } from '@/types'

type Metadata = Record<string, unknown>

type DetailRow = {
  key: string
  label: string
  oldValue?: string
  newValue: string
}

const props = defineProps<{
  message: SupportTicketMessage
}>()

const { t } = useI18n()

const rows = computed(() => {
  const metadata = props.message.metadata
  if (!isMetadata(metadata)) return []
  if (isInvoiceIssuedMessage(props.message)) return invoiceRows(metadata)
  if (!isGroupChangedMessage(props.message)) return []

  const result: DetailRow[] = []
  appendArrayChange(result, metadata, 'allowedGroups', 'old_allowed_groups', 'new_allowed_groups', formatGroupList)
  appendNumberChange(result, metadata, 'rateMultiplier', 'old_rate_multiplier', 'new_rate_multiplier', formatMultiplier)
  appendNumberChange(result, metadata, 'imageRateMultiplier', 'old_image_rate_multiplier', 'new_image_rate_multiplier', formatMultiplier)
  appendNumberChange(result, metadata, 'rpmLimit', 'old_rpm_limit', 'new_rpm_limit', formatLimit)
  appendGroupSwitch(result, metadata)
  appendGroupRateChanges(result, metadata)
  appendRPMOverrideChanges(result, metadata)
  return result
})

function isGroupChangedMessage(message: SupportTicketMessage): boolean {
  return message.sender_type === 'system'
    && (message.event_type === 'group_changed' || message.metadata?.action_type === 'group_changed')
}

function isInvoiceIssuedMessage(message: SupportTicketMessage): boolean {
  return message.sender_type === 'system'
    && (message.event_type === 'invoice_issued' || message.metadata?.action_type === 'invoice_issued')
}

function invoiceRows(metadata: Metadata): DetailRow[] {
  const result: DetailRow[] = []
  const amount = numberValue(metadata.amount)
  const currency = stringValue(metadata.currency)
  const invoiceNo = stringValue(metadata.invoice_no)
  const fileName = stringValue(metadata.file_name)
  if (amount != null) {
    result.push({ key: 'invoiceAmount', label: label('invoiceAmount'), newValue: `${formatNumber(amount)} ${currency || 'CNY'}` })
  }
  if (invoiceNo) {
    result.push({ key: 'invoiceNo', label: label('invoiceNo'), newValue: invoiceNo })
  }
  if (fileName) {
    result.push({ key: 'invoiceFile', label: label('invoiceFile'), newValue: fileName })
  }
  return result
}

function isMetadata(value: unknown): value is Metadata {
  return !!value && typeof value === 'object' && !Array.isArray(value)
}

function stringValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function label(key: string, params?: Record<string, unknown>): string {
  return params ? t(`tickets.metadata.${key}`, params) : t(`tickets.metadata.${key}`)
}

function appendArrayChange(
  rows: DetailRow[],
  metadata: Metadata,
  key: string,
  oldKey: string,
  newKey: string,
  formatter: (value: unknown) => string,
) {
  if (!(oldKey in metadata) && !(newKey in metadata)) return
  const oldValue = oldKey in metadata ? formatter(metadata[oldKey]) : undefined
  const newValue = formatter(metadata[newKey])
  if (oldValue != null && oldValue === newValue) return
  rows.push({
    key,
    label: label(key),
    oldValue,
    newValue,
  })
}

function appendNumberChange(
  rows: DetailRow[],
  metadata: Metadata,
  key: string,
  oldKey: string,
  newKey: string,
  formatter: (value: number) => string,
) {
  const oldValue = numberValue(metadata[oldKey])
  const newValue = numberValue(metadata[newKey])
  if (oldValue == null && newValue == null) return
  if (oldValue != null && newValue != null && oldValue === newValue) return
  rows.push({
    key,
    label: label(key),
    oldValue: oldValue == null ? undefined : formatter(oldValue),
    newValue: newValue == null ? label('unchanged') : formatter(newValue),
  })
}

function appendGroupSwitch(rows: DetailRow[], metadata: Metadata) {
  const oldGroupID = numberValue(metadata.old_group_id)
  const newGroupID = numberValue(metadata.new_group_id)
  if (oldGroupID == null && newGroupID == null) return
  if (oldGroupID != null && newGroupID != null && oldGroupID === newGroupID) return
  rows.push({
    key: 'groupSwitch',
    label: label('groupSwitch'),
    oldValue: oldGroupID == null ? undefined : formatGroupID(oldGroupID),
    newValue: newGroupID == null ? label('cleared') : formatGroupID(newGroupID),
  })
}

function appendGroupRateChanges(rows: DetailRow[], metadata: Metadata) {
  for (const item of metadataList(metadata.group_rate_changes)) {
    const groupID = numberValue(item.group_id)
    const oldValue = numberValue(item.old_rate_multiplier)
    const newValue = numberValue(item.new_rate_multiplier)
    const cleared = item.cleared === true
    if (oldValue == null && newValue == null && !cleared) continue
    rows.push({
      key: `groupRate:${groupID ?? rows.length}`,
      label: groupID == null ? label('groupRate') : label('groupRateWithId', { id: formatID(groupID) }),
      oldValue: oldValue == null ? undefined : formatMultiplier(oldValue),
      newValue: cleared ? label('useDefaultRate') : newValue == null ? label('unchanged') : formatMultiplier(newValue),
    })
  }
}

function appendRPMOverrideChanges(rows: DetailRow[], metadata: Metadata) {
  for (const item of metadataList(metadata.rpm_override_changes)) {
    const groupID = numberValue(item.group_id)
    const oldValue = numberValue(item.old_rpm_override)
    const newValue = numberValue(item.new_rpm_override)
    const cleared = item.cleared === true
    if (oldValue == null && newValue == null && !cleared) continue
    rows.push({
      key: `rpmOverride:${groupID ?? rows.length}`,
      label: groupID == null ? label('rpmOverride') : label('rpmOverrideWithId', { id: formatID(groupID) }),
      oldValue: oldValue == null ? undefined : formatLimit(oldValue),
      newValue: cleared ? label('useDefaultLimit') : newValue == null ? label('unchanged') : formatLimit(newValue),
    })
  }
}

function metadataList(value: unknown): Metadata[] {
  if (!Array.isArray(value)) return []
  return value.filter(isMetadata)
}

function numberValue(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

function formatMultiplier(value: number): string {
  return `${formatNumber(value)}x`
}

function formatLimit(value: number): string {
  return value <= 0 ? label('unlimited') : formatNumber(value)
}

function formatGroupList(value: unknown): string {
  if (!Array.isArray(value) || value.length === 0) return label('none')
  const ids = value
    .map(numberValue)
    .filter((id): id is number => id != null)
  if (ids.length === 0) return label('none')
  return ids.map(formatGroupID).join(', ')
}

function formatGroupID(id: number): string {
  return label('groupId', { id: formatID(id) })
}

function formatID(id: number): string {
  return String(Math.trunc(id))
}

function formatNumber(value: number): string {
  return Number.isInteger(value) ? String(value) : String(Number(value.toFixed(4)))
}
</script>

<style scoped>
.system-metadata-details {
  border-top: 1px solid rgba(146, 64, 14, 0.18);
  padding-top: 0.75rem;
}

.metadata-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.25rem;
  align-items: start;
  border-radius: 0.5rem;
  background: rgba(255, 255, 255, 0.5);
  padding: 0.5rem 0.625rem;
}

.metadata-label {
  min-width: 0;
  color: rgb(146 64 14);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.25rem;
}

.metadata-value {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  gap: 0.375rem;
  color: rgb(120 53 15);
  font-size: 0.8125rem;
  line-height: 1.25rem;
  word-break: break-word;
  overflow-wrap: anywhere;
}

.metadata-old {
  opacity: 0.72;
  text-decoration: line-through;
}

.metadata-arrow {
  opacity: 0.65;
}

.dark .system-metadata-details {
  border-top-color: rgba(251, 191, 36, 0.22);
}

.dark .metadata-row {
  background: rgba(120, 53, 15, 0.28);
}

.dark .metadata-label {
  color: rgb(253 230 138);
}

.dark .metadata-value {
  color: rgb(254 243 199);
}

@media (min-width: 640px) {
  .metadata-row {
    grid-template-columns: minmax(6rem, 0.45fr) minmax(0, 1fr);
    gap: 0.75rem;
  }
}
</style>
