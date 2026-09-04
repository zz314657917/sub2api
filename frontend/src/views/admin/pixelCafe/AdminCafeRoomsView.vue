<template>
  <AppLayout>
    <div class="space-y-4">
      <nav
        class="flex flex-wrap items-center gap-2 border-b border-gray-200 pb-3 dark:border-dark-700"
        aria-label="像素网吧管理视图"
        data-testid="cafe-workspace-tabs"
      >
        <button
          type="button"
          class="btn btn-sm"
          :class="activeWorkspace === 'rooms' ? 'btn-primary' : 'btn-ghost'"
          :aria-pressed="activeWorkspace === 'rooms'"
          @click="switchWorkspace('rooms')"
        >
          <Icon name="home" size="sm" class="mr-1" />
          房间管理
        </button>
        <button
          type="button"
          class="btn btn-sm"
          :class="activeWorkspace === 'rounds' ? 'btn-primary' : 'btn-ghost'"
          :aria-pressed="activeWorkspace === 'rounds'"
          @click="switchWorkspace('rounds')"
        >
          <Icon name="clipboard" size="sm" class="mr-1" />
          团次处理
          <span
            v-if="pendingRounds.length > 0"
            class="ml-1 rounded-full bg-amber-100 px-1.5 py-0.5 text-[11px] font-semibold text-amber-800 dark:bg-amber-900/40 dark:text-amber-200"
          >
            {{ pendingRounds.length }}
          </span>
        </button>
      </nav>

      <TablePageLayout v-if="activeWorkspace === 'rooms'">
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="min-w-56 flex-1 sm:max-w-72">
            <input
              v-model="search"
              class="input"
              type="search"
              :placeholder="t('admin.pixelCafe.searchPlaceholder')"
              @input="scheduleReload"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="statusOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <Select
            v-model="filters.zone"
            :options="zoneOptions"
            class="w-40"
            @change="resetAndLoad"
          />
          <div class="ml-auto flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="btn btn-secondary"
              :disabled="loading"
              :title="t('admin.pixelCafe.refresh')"
              @click="loadRooms"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              <span class="sr-only">{{ t('admin.pixelCafe.refresh') }}</span>
            </button>
            <button type="button" class="btn btn-secondary" @click="openLayoutDialog">
              <Icon name="edit" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.layout.open') }}
            </button>
            <button type="button" class="btn btn-secondary" @click="openBulkDialog">
              <Icon name="copy" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.bulkCreate') }}
            </button>
            <button type="button" class="btn btn-secondary" :disabled="quotaResetting === 'all'" @click="askResetAllQuotas">
              <Icon v-if="quotaResetting === 'all'" name="refresh" size="md" class="mr-1 animate-spin" />
              <Icon v-else name="refresh" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.quotaReset.allButton') }}
            </button>
            <button type="button" class="btn btn-primary" @click="openCreateDialog">
              <Icon name="plus" size="md" class="mr-1" />
              {{ t('admin.pixelCafe.createRoom') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="rooms"
          :loading="loading"
          :sticky-first-column="true"
          :sticky-actions-column="true"
          :expandable-actions="false"
          row-key="id"
        >
          <template #cell-room="{ row }">
            <div class="min-w-44">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">{{ row.code }} · #{{ row.id }}</div>
            </div>
          </template>
          <template #cell-zone="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ row.zone_key || 'featured' }}</span>
          </template>
          <template #cell-plan="{ row }">
            <div class="min-w-44">
              <div class="font-medium text-gray-800 dark:text-gray-200">ChatGPT {{ row.plan?.subscription_tier === 'pro' ? 'Pro' : 'Plus' }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">
                {{ row.plan?.total_shares || '-' }} 份 · {{ formatPrice(row.plan?.price_per_share) }} · {{ row.plan?.current_round_status || '无进行中团次' }}
              </div>
            </div>
          </template>
          <template #cell-account>
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.pixelCafe.accountDeferred') }}</span>
          </template>
          <template #cell-status="{ row }">
            <span class="status-badge" :class="statusClass(row.status)">
              {{ t(`admin.pixelCafe.status.${row.status}`) }}
            </span>
          </template>
          <template #cell-sort_order="{ row }">
            <span class="font-mono text-sm text-gray-700 dark:text-gray-300">{{ row.sort_order }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex min-w-56 flex-wrap items-center gap-2">
              <button type="button" class="btn btn-ghost btn-sm" @click="openEditDialog(row)">
                <Icon name="edit" size="sm" class="mr-1" />
                {{ t('admin.pixelCafe.actions.edit') }}
              </button>
              <button
                type="button"
                class="btn btn-secondary btn-sm"
                :disabled="!canOperateRound(row) || roundActionBusy(row)"
                @click="handleRoundAction(row)"
              >
                <Icon name="play" size="sm" class="mr-1" />
                {{ roundActionLabel(row) }}
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-sm text-red-600 hover:text-red-700 dark:text-red-300"
                :disabled="row.status === 'enabled' || deletingId === row.id"
                @click="askDelete(row)"
              >
                <Icon name="trash" size="sm" class="mr-1" />
                {{ t('admin.pixelCafe.actions.delete') }}
              </button>
              <button
                type="button"
                class="btn btn-ghost btn-sm"
                :disabled="quotaResetting === row.id"
                @click="askResetRoomQuotas(row)"
              >
                <Icon v-if="quotaResetting === row.id" name="refresh" size="sm" class="mr-1 animate-spin" />
                <Icon v-else name="refresh" size="sm" class="mr-1" />
                {{ t('admin.pixelCafe.quotaReset.roomButton') }}
              </button>
            </div>
          </template>
          <template #empty>
            <EmptyState
              :title="t('admin.pixelCafe.noRooms')"
              :description="loadError || t('admin.pixelCafe.noRooms')"
              :action-text="t('admin.pixelCafe.createRoom')"
              @action="openCreateDialog"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="changePage"
          @update:page-size="changePageSize"
        />
      </template>
      </TablePageLayout>

      <section v-else class="space-y-4" data-testid="cafe-round-workspace">
        <section class="rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-900 dark:bg-amber-950/20" data-testid="cafe-pending-fulfillment">
          <div class="mb-3 flex flex-wrap items-center gap-3">
            <div>
              <h2 class="font-semibold">{{ t('admin.pixelCafe.pending.title') }}</h2>
              <p class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.pixelCafe.pending.description') }}</p>
            </div>
            <input v-model="pendingSearch" class="input ml-auto max-w-xs" type="search" :placeholder="t('admin.pixelCafe.pending.search')" @change="loadPendingRounds" />
          </div>
          <p v-if="pendingLoading" class="text-sm">{{ t('admin.pixelCafe.pending.loading') }}</p>
          <p v-else-if="pendingRounds.length === 0" class="text-sm text-gray-600 dark:text-dark-300">{{ t('admin.pixelCafe.pending.empty') }}</p>
          <div v-else class="space-y-2">
            <div v-for="round in pendingRounds" :key="round.id" class="flex flex-wrap items-center gap-3 rounded border border-amber-200 bg-white p-3 dark:border-amber-900 dark:bg-dark-900">
              <span class="font-medium">{{ round.room_code }} · {{ round.room_name }}</span>
              <span>ChatGPT {{ round.subscription_tier === 'pro' ? 'Pro' : 'Plus' }}</span>
              <span>{{ round.paid_shares }}/{{ round.total_shares }} 份 · {{ round.joined_buyers }}/{{ round.max_buyers }} 人</span>
              <button type="button" class="btn btn-secondary btn-sm ml-auto" @click="openAssignDialog(round)">{{ t('admin.pixelCafe.pending.assign') }}</button>
            </div>
          </div>
        </section>

        <AdminGroupBuyView embedded rounds-only />
      </section>
    </div>

    <BaseDialog
      :show="layoutDialogOpen"
      :title="t('admin.pixelCafe.layout.title')"
      width="wide"
      @close="closeLayoutDialog"
    >
      <div v-if="layoutLoading" class="py-10 text-center text-sm text-gray-500 dark:text-dark-400">
        {{ t('admin.pixelCafe.layout.loading') }}
      </div>
      <CafeWorkstationLayoutEditor v-else v-model="workstationLayoutDraft" />
      <template #footer>
        <div class="flex w-full flex-wrap items-center justify-between gap-3">
          <button type="button" class="btn btn-ghost" :disabled="layoutLoading || layoutSaving" @click="resetWorkstationLayout">
            {{ t('admin.pixelCafe.layout.reset') }}
          </button>
          <div class="flex gap-3">
            <button type="button" class="btn btn-secondary" :disabled="layoutSaving" @click="closeLayoutDialog">
              {{ t('common.cancel') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="layoutLoading || layoutSaving" @click="saveWorkstationLayout">
              <Icon v-if="layoutSaving" name="refresh" size="sm" class="mr-1 animate-spin" />
              {{ layoutSaving ? t('admin.pixelCafe.layout.saving') : t('admin.pixelCafe.layout.save') }}
            </button>
          </div>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog
      :show="roomDialogOpen"
      :title="editingRoom ? t('admin.pixelCafe.form.editTitle') : t('admin.pixelCafe.form.createTitle')"
      width="wide"
      @close="closeRoomDialog"
    >
      <form id="cafe-room-form" class="space-y-4" @submit.prevent="saveRoom">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="field">
            <span class="input-label">房间编号（自动生成）</span>
            <input v-model.trim="roomForm.code" class="input" maxlength="64" readonly placeholder="保存后自动生成" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.name') }}</span>
            <input v-model.trim="roomForm.name" class="input" required maxlength="120" />
          </label>
          <label class="field sm:col-span-2">
            <span class="input-label">房间说明</span>
            <textarea v-model.trim="roomForm.description" class="input min-h-20" maxlength="2000" :disabled="commercialLocked" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.zone') }}</span>
            <input v-model.trim="roomForm.zone_key" class="input" maxlength="32" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.theme') }}</span>
            <input v-model.trim="roomForm.theme_key" class="input" maxlength="64" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.sceneSlot') }}</span>
            <input v-model.trim="roomForm.scene_slot_key" class="input" maxlength="120" />
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.status') }}</span>
            <select v-model="roomForm.status" class="input" :disabled="commercialLocked">
              <option v-for="option in statusOptions.slice(1)" :key="String(option.value)" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.sortOrder') }}</span>
            <input v-model.number="roomForm.sort_order" class="input" type="number" step="1" />
            <span class="field-hint">{{ t('admin.pixelCafe.form.sortOrderHint') }}</span>
          </label>
        </div>
        <section class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-dark-700" data-testid="room-owned-plan-fields">
          <div>
            <h3 class="font-semibold text-gray-900 dark:text-white">套餐与份额</h3>
            <p class="text-xs text-gray-500 dark:text-dark-400">每个房间自动创建专属计划；有进行中团次时商业配置会锁定。</p>
          </div>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <label class="field"><span class="input-label">账号套餐</span><select v-model="roomForm.plan!.subscription_tier" class="input" :disabled="commercialLocked"><option value="plus">ChatGPT Plus</option><option value="pro">ChatGPT Pro</option></select></label>
            <label class="field"><span class="input-label">总份额</span><input v-model.number="roomForm.plan!.total_shares" class="input" type="number" min="1" max="10" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">最多参与人数</span><input v-model.number="roomForm.plan!.max_buyers" class="input" type="number" min="1" :max="roomForm.plan!.total_shares" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">单用户最大份额</span><input v-model.number="roomForm.plan!.max_shares_per_user" class="input" type="number" min="1" :max="roomForm.plan!.total_shares" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">每份价格</span><input v-model.number="roomForm.plan!.price_per_share" class="input" type="number" min="0.01" step="0.01" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">价格展示文案</span><input v-model.trim="roomForm.plan!.price_label" class="input" maxlength="120" :disabled="commercialLocked" placeholder="可留空，自动显示价格" /></label>
            <label class="field"><span class="input-label">拼团截止（分钟）</span><input v-model.number="roomForm.plan!.timeout_minutes" class="input" type="number" min="1" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">成团后配号时限（分钟）</span><input v-model.number="roomForm.plan!.fulfillment_timeout_minutes" class="input" type="number" min="1" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">有效期（天）</span><input v-model.number="roomForm.plan!.validity_days" class="input" type="number" min="1" :disabled="commercialLocked" /></label>
            <div class="field sm:col-span-2 lg:col-span-3"><span class="input-label">托管订阅分组</span><div class="input bg-gray-50 text-gray-600 dark:bg-dark-800 dark:text-gray-300">自动使用系统默认的网吧托管订阅分组</div><span class="field-hint">该分组由系统自动创建和维护，不需要手动选择。</span></div>
          </div>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <label class="field"><span class="input-label">每份 Key 总额度</span><input v-model.number="roomForm.plan!.room_key_quota_usd" class="input" type="number" min="0" step="0.01" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">每份 5H 限额</span><input v-model.number="roomForm.plan!.room_key_rate_limit_5h" class="input" type="number" min="0" step="0.01" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">每份 1D 限额</span><input v-model.number="roomForm.plan!.room_key_rate_limit_1d" class="input" type="number" min="0" step="0.01" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">每份 7D 限额</span><input v-model.number="roomForm.plan!.room_key_rate_limit_7d" class="input" type="number" min="0" step="0.01" :disabled="commercialLocked" /></label>
            <label class="field sm:col-span-2"><span class="input-label">额度展示文案</span><input v-model.trim="roomForm.plan!.quota_per_share_label" class="input" maxlength="255" :disabled="commercialLocked" /></label>
            <label class="field"><span class="input-label">退款方式</span><select v-model="roomForm.plan!.refund_mode" class="input" :disabled="commercialLocked"><option value="balance_credit">退回余额</option><option value="provider_refund">原路退款</option></select></label>
            <label class="field sm:col-span-2 lg:col-span-4"><span class="input-label">用户协议</span><textarea v-model.trim="roomForm.plan!.agreement_text" class="input min-h-20" :disabled="commercialLocked" maxlength="4000" /></label>
          </div>
        </section>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeRoomDialog">{{ t('common.cancel') }}</button>
          <button type="submit" form="cafe-room-form" class="btn btn-primary" :disabled="saving || !canSaveRoom">
            <Icon v-if="saving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ saving ? t('admin.pixelCafe.form.saving') : t('admin.pixelCafe.form.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BaseDialog :show="Boolean(assigningRound)" :title="t('admin.pixelCafe.pending.assignTitle')" @close="assigningRound = null">
      <div class="space-y-3"><input v-model="accountSearch" class="input" type="search" :placeholder="t('admin.pixelCafe.pending.accountSearch')" @change="loadRoundAccountOptions" /><p v-if="accountLoading" class="text-sm">{{ t('admin.pixelCafe.pending.loading') }}</p><label v-for="account in roundAccountOptions" :key="account.id" class="flex cursor-pointer items-center gap-3 rounded border p-2"><input v-model="selectedAccountID" type="radio" :value="account.id" /><span>{{ account.name }} · {{ account.platform }} · {{ account.plan_type || '-' }}<small v-if="account.email_masked"> · {{ account.email_masked }}</small></span></label><p v-if="!accountLoading && roundAccountOptions.length === 0" class="text-sm">{{ t('admin.pixelCafe.pending.noAccount') }}</p></div>
      <template #footer><div class="flex justify-end gap-3"><button type="button" class="btn btn-secondary" @click="assigningRound = null">{{ t('common.cancel') }}</button><button type="button" class="btn btn-primary" :disabled="!selectedAccountID || assigning" @click="assignAccount">{{ assigning ? t('admin.pixelCafe.pending.assigning') : t('admin.pixelCafe.pending.assign') }}</button></div></template>
    </BaseDialog>

    <BaseDialog
      :show="bulkDialogOpen"
      :title="t('admin.pixelCafe.bulk.title')"
      width="wide"
      @close="closeBulkDialog"
    >
      <form id="cafe-room-bulk-form" class="space-y-4" @submit.prevent="submitBulkCreate">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.bulk.quantity') }}</span>
            <input v-model.number="bulkForm.quantity" class="input" type="number" min="1" max="100" />
          </label>
          <label class="field"><span class="input-label">账号套餐</span><select v-model="bulkForm.plan_template.subscription_tier" class="input"><option value="plus">ChatGPT Plus</option><option value="pro">ChatGPT Pro</option></select></label>
          <label class="field">
            <span class="input-label">{{ t('admin.pixelCafe.form.zone') }}</span>
            <input v-model.trim="bulkForm.zone_key" class="input" maxlength="32" />
          </label>
          <label class="field sm:col-span-2">
            <span class="input-label">{{ t('admin.pixelCafe.form.theme') }}</span>
            <input v-model.trim="bulkForm.theme_key" class="input" maxlength="64" />
          </label>
          <label class="field"><span class="input-label">总份额</span><input v-model.number="bulkForm.plan_template.total_shares" class="input" type="number" min="1" max="10" /></label>
          <label class="field"><span class="input-label">最多参与人数</span><input v-model.number="bulkForm.plan_template.max_buyers" class="input" type="number" min="1" :max="bulkForm.plan_template.total_shares" /></label>
          <label class="field"><span class="input-label">单用户最大份额</span><input v-model.number="bulkForm.plan_template.max_shares_per_user" class="input" type="number" min="1" :max="bulkForm.plan_template.total_shares" /></label>
          <label class="field"><span class="input-label">每份价格</span><input v-model.number="bulkForm.plan_template.price_per_share" class="input" type="number" min="0.01" step="0.01" /></label>
          <label class="field"><span class="input-label">有效期（天）</span><input v-model.number="bulkForm.plan_template.validity_days" class="input" type="number" min="1" /></label>
          <label class="field"><span class="input-label">拼团截止（分钟）</span><input v-model.number="bulkForm.plan_template.timeout_minutes" class="input" type="number" min="1" /></label>
          <label class="field"><span class="input-label">配号时限（分钟）</span><input v-model.number="bulkForm.plan_template.fulfillment_timeout_minutes" class="input" type="number" min="1" /></label>
          <div class="field sm:col-span-2"><span class="input-label">托管订阅分组</span><div class="input bg-gray-50 text-gray-600 dark:bg-dark-800 dark:text-gray-300">自动使用系统默认的网吧托管订阅分组</div></div>
          <label class="field"><span class="input-label">每份 Key 总额度</span><input v-model.number="bulkForm.plan_template.room_key_quota_usd" class="input" type="number" min="0" step="0.01" /></label>
          <label class="field"><span class="input-label">每份 5H 限额</span><input v-model.number="bulkForm.plan_template.room_key_rate_limit_5h" class="input" type="number" min="0" step="0.01" /></label>
          <label class="field"><span class="input-label">每份 1D 限额</span><input v-model.number="bulkForm.plan_template.room_key_rate_limit_1d" class="input" type="number" min="0" step="0.01" /></label>
          <label class="field"><span class="input-label">每份 7D 限额</span><input v-model.number="bulkForm.plan_template.room_key_rate_limit_7d" class="input" type="number" min="0" step="0.01" /></label>
          <label class="field"><span class="input-label">退款方式</span><select v-model="bulkForm.plan_template.refund_mode" class="input"><option value="balance_credit">退回余额</option><option value="provider_refund">原路退款</option></select></label>
          <label class="field sm:col-span-2"><span class="input-label">价格展示文案</span><input v-model.trim="bulkForm.plan_template.price_label" class="input" maxlength="120" /></label>
          <label class="field sm:col-span-2"><span class="input-label">额度展示文案</span><input v-model.trim="bulkForm.plan_template.quota_per_share_label" class="input" maxlength="255" /></label>
          <label class="field sm:col-span-2"><span class="input-label">用户协议</span><textarea v-model.trim="bulkForm.plan_template.agreement_text" class="input min-h-20" maxlength="4000" /></label>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="bulkForm.create_open_round" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600" />
          {{ t('admin.pixelCafe.bulk.createOpenRound') }}
        </label>
        <div v-if="bulkResult" class="space-y-3 rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900/50">
          <div class="flex flex-wrap gap-3 text-sm">
            <span class="text-emerald-700 dark:text-emerald-300">{{ t('admin.pixelCafe.bulk.created', { count: bulkResult.created.length }) }}</span>
            <span class="text-red-700 dark:text-red-300">{{ t('admin.pixelCafe.bulk.failed', { count: bulkResult.failed.length }) }}</span>
          </div>
          <ul v-if="bulkResult.failed.length" class="space-y-1 text-sm text-red-700 dark:text-red-300">
            <li v-for="failure in bulkResult.failed" :key="`${failure.index}-${failure.error_code}`">
              #{{ failure.index ?? '-' }} · {{ failure.error_code }} · {{ failure.message }}
            </li>
          </ul>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeBulkDialog">{{ t('common.close') }}</button>
          <button type="submit" form="cafe-room-bulk-form" class="btn btn-primary" :disabled="bulkSaving || !canSaveBulk">
            <Icon v-if="bulkSaving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ bulkSaving ? t('admin.pixelCafe.bulk.submitting') : t('admin.pixelCafe.bulk.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <ConfirmDialog
      :show="Boolean(roomToDelete)"
      :title="t('admin.pixelCafe.confirmDeleteTitle')"
      :message="roomToDelete ? t('admin.pixelCafe.confirmDeleteMessage', { name: roomToDelete.name }) : ''"
      :danger="true"
      @cancel="roomToDelete = null"
      @confirm="deleteRoom"
    />

    <ConfirmDialog
      :show="Boolean(quotaResetScope)"
      :title="t('admin.pixelCafe.quotaReset.confirmTitle')"
      :message="quotaResetConfirmMessage"
      @cancel="closeQuotaResetConfirm"
      @confirm="confirmQuotaReset"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { Column } from '@/components/common/types'
import type { CafeRoom, CafeRoomBulkResult, CafeRoomInput, CafeRoomPlanInput, CafeRoomStatus, CafeWorkstationPosition } from '@/types/pixelCafe'
import { createCafeWorkstationLayout, resolveCafeWorkstationLayout } from '@/features/pixelCafe/renderer/sceneLayout'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminGroupBuyView from '@/views/admin/group-buy/AdminGroupBuyView.vue'
import CafeWorkstationLayoutEditor from './components/CafeWorkstationLayoutEditor.vue'
import type { CafePendingRound, CafeRoomAccountOption } from '@/api/admin/cafeRooms'

const { t } = useI18n()
const appStore = useAppStore()

const activeWorkspace = ref<'rooms' | 'rounds'>('rooms')
const rooms = ref<CafeRoom[]>([])
const pendingRounds = ref<CafePendingRound[]>([])
const pendingSearch = ref('')
const pendingLoading = ref(false)
const assigningRound = ref<CafePendingRound | null>(null)
const roundAccountOptions = ref<CafeRoomAccountOption[]>([])
const accountSearch = ref('')
const selectedAccountID = ref(0)
const accountLoading = ref(false)
const assigning = ref(false)
const loading = ref(false)
const loadError = ref('')
const search = ref('')
const filters = reactive({ status: '', zone: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0, pages: 1 })

const roomDialogOpen = ref(false)
const bulkDialogOpen = ref(false)
const editingRoom = ref<CafeRoom | null>(null)
const saving = ref(false)
const bulkSaving = ref(false)
const openingRoundId = ref<number | null>(null)
const pausingRoundId = ref<number | null>(null)
const deletingId = ref<number | null>(null)
const roomToDelete = ref<CafeRoom | null>(null)
const quotaResetScope = ref<'room' | 'all' | null>(null)
const quotaResetRoom = ref<CafeRoom | null>(null)
const quotaResetting = ref<number | 'all' | null>(null)
const bulkResult = ref<CafeRoomBulkResult | null>(null)
const layoutDialogOpen = ref(false)
const layoutLoading = ref(false)
const layoutSaving = ref(false)
const workstationLayoutDraft = ref<CafeWorkstationPosition[]>(resolveCafeWorkstationLayout())

function defaultPlanInput(): CafeRoomPlanInput {
  return {
    subscription_tier: 'plus', total_shares: 10, max_buyers: 4, max_shares_per_user: 10,
    price_per_share: 1, price_label: '', quota_per_share_label: '', timeout_minutes: 1440,
    fulfillment_timeout_minutes: 1440, validity_days: 30, target_group_id: 0,
    room_key_quota_usd: 0, room_key_rate_limit_5h: 0, room_key_rate_limit_1d: 0, room_key_rate_limit_7d: 0,
    refund_mode: 'balance_credit', agreement_text: '',
  }
}

const roomForm = reactive<CafeRoomInput>({
  code: '',
  name: '',
  description: '',
  plan: defaultPlanInput(),
  zone_key: 'featured',
  theme_key: 'warm_wood',
  scene_slot_key: '',
  status: 'draft',
  featured: false,
  sort_order: 0,
})

const bulkForm = reactive({
  plan_template: defaultPlanInput(),
  quantity: 1,
  zone_key: 'featured',
  theme_key: 'warm_wood',
  create_open_round: false,
})

const columns = computed<Column[]>(() => [
  { key: 'room', label: t('admin.pixelCafe.columns.room'), sortable: true },
  { key: 'zone', label: t('admin.pixelCafe.columns.zone'), sortable: true },
  { key: 'plan', label: t('admin.pixelCafe.columns.plan') },
  { key: 'account', label: t('admin.pixelCafe.columns.account') },
  { key: 'status', label: t('admin.pixelCafe.columns.status'), sortable: true },
  { key: 'sort_order', label: t('admin.pixelCafe.columns.sortOrder') },
  { key: 'actions', label: t('admin.pixelCafe.columns.actions') },
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.pixelCafe.allStatus') },
  ...(['draft', 'enabled', 'maintenance', 'disabled'] as CafeRoomStatus[]).map((status) => ({
    value: status,
    label: t(`admin.pixelCafe.status.${status}`),
  })),
])

const zoneOptions = computed(() => [
  { value: '', label: t('admin.pixelCafe.allZones') },
  ...Array.from(new Set(rooms.value.map((room) => room.zone_key).filter(Boolean))).map((zone) => ({ value: zone, label: zone })),
])

const commercialLocked = computed(() => Boolean(editingRoom.value?.plan?.current_round_status))
const canSaveRoom = computed(() => {
  const plan = roomForm.plan
  return Boolean(roomForm.name.trim() && plan && plan.price_per_share > 0 && plan.total_shares >= 1 && plan.total_shares <= 10 && plan.max_buyers >= 1 && plan.max_buyers <= plan.total_shares && plan.max_shares_per_user >= 1 && plan.max_shares_per_user <= plan.total_shares)
})
const canSaveBulk = computed(() => {
  const plan = bulkForm.plan_template
  return bulkForm.quantity >= 1 && bulkForm.quantity <= 100 && plan.price_per_share > 0 && plan.total_shares >= 1 && plan.total_shares <= 10 && plan.max_buyers >= 1 && plan.max_buyers <= plan.total_shares && plan.max_shares_per_user >= 1 && plan.max_shares_per_user <= plan.total_shares
})
function resetRoomForm() {
  Object.assign(roomForm, {
    code: '', name: '', description: '', plan_id: undefined, plan: defaultPlanInput(),
    zone_key: 'featured', theme_key: 'warm_wood', scene_slot_key: '', status: 'draft', featured: false, sort_order: 0,
  })
}

function openCreateDialog() {
  editingRoom.value = null
  resetRoomForm()
  roomDialogOpen.value = true
}

function openEditDialog(room: CafeRoom) {
  editingRoom.value = room
  Object.assign(roomForm, {
    code: room.code, name: room.name, description: room.description || room.plan?.description || '', plan_id: undefined,
    plan: room.plan ? {
      subscription_tier: room.plan.subscription_tier || 'plus', total_shares: room.plan.total_shares,
      max_buyers: room.plan.max_buyers || Math.min(room.plan.total_shares, 4), max_shares_per_user: room.plan.max_shares_per_user || room.plan.total_shares,
      price_per_share: room.plan.price_per_share, price_label: room.plan.price_label || '', quota_per_share_label: room.plan.quota_per_share_label || '',
      timeout_minutes: room.plan.timeout_minutes, fulfillment_timeout_minutes: room.plan.fulfillment_timeout_minutes || 1440,
      validity_days: room.plan.validity_days, target_group_id: room.plan.target_group_id,
      room_key_quota_usd: room.plan.room_key_quota_usd || 0, room_key_rate_limit_5h: room.plan.room_key_rate_limit_5h || 0,
      room_key_rate_limit_1d: room.plan.room_key_rate_limit_1d || 0, room_key_rate_limit_7d: room.plan.room_key_rate_limit_7d || 0,
      refund_mode: room.plan.refund_mode || 'balance_credit', agreement_text: room.plan.agreement_text || '',
    } : defaultPlanInput(),
    zone_key: room.zone_key, theme_key: room.theme_key, scene_slot_key: room.scene_slot_key,
    status: room.status, featured: room.featured, sort_order: room.sort_order,
  })
  roomDialogOpen.value = true
}

function closeRoomDialog() {
  if (saving.value) return
  roomDialogOpen.value = false
}

function openBulkDialog() {
  bulkResult.value = null
  bulkForm.plan_template = defaultPlanInput()
  bulkForm.quantity = 1
  bulkDialogOpen.value = true
}

function closeBulkDialog() {
  if (bulkSaving.value) return
  bulkDialogOpen.value = false
}

async function openLayoutDialog() {
  layoutDialogOpen.value = true
  layoutLoading.value = true
  try {
    const response = await adminAPI.cafeRooms.getWorkstationLayout()
    workstationLayoutDraft.value = resolveCafeWorkstationLayout(response.data)
  } catch (error) {
    workstationLayoutDraft.value = resolveCafeWorkstationLayout()
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.layout.loadError')))
  } finally {
    layoutLoading.value = false
  }
}

function closeLayoutDialog() {
  if (layoutSaving.value) return
  layoutDialogOpen.value = false
}

function resetWorkstationLayout() {
  workstationLayoutDraft.value = createCafeWorkstationLayout(workstationLayoutDraft.value.length)
}

async function saveWorkstationLayout() {
  layoutSaving.value = true
  try {
    const response = await adminAPI.cafeRooms.updateWorkstationLayout(workstationLayoutDraft.value)
    workstationLayoutDraft.value = resolveCafeWorkstationLayout(response.data)
    layoutDialogOpen.value = false
    appStore.showSuccess(t('admin.pixelCafe.success.layoutSaved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.layout.saveError')))
  } finally {
    layoutSaving.value = false
  }
}

function statusClass(status: string) {
  return {
    draft: 'status-badge-muted',
    enabled: 'status-badge-success',
    maintenance: 'status-badge-warning',
    disabled: 'status-badge-danger',
  }[status] || 'status-badge-muted'
}

function formatPrice(value?: number) {
  return value != null ? `¥${Number(value).toFixed(2)}/份` : '-'
}

function switchWorkspace(workspace: 'rooms' | 'rounds') {
  activeWorkspace.value = workspace
  if (workspace === 'rounds') void loadPendingRounds()
}

let searchTimer: number | null = null
function scheduleReload() {
  if (searchTimer) window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(resetAndLoad, 300)
}

function resetAndLoad() {
  pagination.page = 1
  void loadRooms()
}

async function loadRooms() {
  loading.value = true
  loadError.value = ''
  try {
    const response = await adminAPI.cafeRooms.list({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      zone: filters.zone || undefined,
      search: search.value.trim() || undefined,
      sort_by: 'sort_order',
      sort_order: 'asc',
    })
    rooms.value = response.data.items
    pagination.total = response.data.total
    pagination.pages = response.data.pages
    pagination.page = response.data.page
    pagination.page_size = response.data.page_size
  } catch (error) {
    loadError.value = extractApiErrorMessage(error, t('admin.pixelCafe.errors.load'))
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

function changePage(page: number) {
  pagination.page = page
  void loadRooms()
}

function changePageSize(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadRooms()
}

async function saveRoom() {
  if (!canSaveRoom.value) return
  saving.value = true
  try {
    const payload = { ...roomForm }
    if (!editingRoom.value) payload.code = ''
    if (!payload.plan_id) delete payload.plan_id
    if (editingRoom.value) {
      await adminAPI.cafeRooms.update(editingRoom.value.id, payload)
      appStore.showSuccess(t('admin.pixelCafe.success.updated'))
    } else {
      await adminAPI.cafeRooms.create(payload)
      appStore.showSuccess(t('admin.pixelCafe.success.created'))
    }
    roomDialogOpen.value = false
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.save')))
  } finally {
    saving.value = false
  }
}

function askDelete(room: CafeRoom) {
  roomToDelete.value = room
}

const quotaResetConfirmMessage = computed(() => {
  if (quotaResetScope.value === 'room' && quotaResetRoom.value) {
    return t('admin.pixelCafe.quotaReset.roomMessage', { name: quotaResetRoom.value.name })
  }
  return t('admin.pixelCafe.quotaReset.allMessage')
})

function askResetRoomQuotas(room: CafeRoom) {
  quotaResetRoom.value = room
  quotaResetScope.value = 'room'
}

function askResetAllQuotas() {
  quotaResetRoom.value = null
  quotaResetScope.value = 'all'
}

function closeQuotaResetConfirm() {
  if (quotaResetting.value !== null) return
  quotaResetScope.value = null
  quotaResetRoom.value = null
}

async function confirmQuotaReset() {
  const scope = quotaResetScope.value
  const room = quotaResetRoom.value
  if (!scope || quotaResetting.value !== null || (scope === 'room' && !room)) return
  quotaResetting.value = scope === 'all' ? 'all' : room!.id
  try {
    const response = scope === 'all'
      ? await adminAPI.cafeRooms.resetAllQuotas()
      : await adminAPI.cafeRooms.resetRoomQuotas(room!.id)
    appStore.showSuccess(t('admin.pixelCafe.quotaReset.success', { count: response.data.affected_keys }))
    quotaResetScope.value = null
    quotaResetRoom.value = null
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.quotaReset.error')))
  } finally {
    quotaResetting.value = null
  }
}

async function deleteRoom() {
  if (!roomToDelete.value) return
  const room = roomToDelete.value
  deletingId.value = room.id
  try {
    await adminAPI.cafeRooms.remove(room.id)
    roomToDelete.value = null
    appStore.showSuccess(t('admin.pixelCafe.success.deleted'))
    if (rooms.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.delete')))
  } finally {
    deletingId.value = null
  }
}

async function openRound(room: CafeRoom) {
  openingRoundId.value = room.id
  try {
    await adminAPI.cafeRooms.openRound(room.id)
    appStore.showSuccess(t('admin.pixelCafe.success.roundOpened'))
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.openRound')))
  } finally {
    openingRoundId.value = null
  }
}

async function pauseRound(room: CafeRoom) {
  pausingRoundId.value = room.id
  try {
    await adminAPI.cafeRooms.pauseRound(room.id)
    appStore.showSuccess(t('admin.pixelCafe.success.roundPaused'))
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.pauseRound')))
  } finally {
    pausingRoundId.value = null
  }
}

function roundStatus(room: CafeRoom): string {
  return room.plan?.current_round_status || ''
}

function canOperateRound(room: CafeRoom): boolean {
  return room.status === 'enabled' && (roundStatus(room) === '' || roundStatus(room) === 'open')
}

function roundActionBusy(room: CafeRoom): boolean {
  return openingRoundId.value === room.id || pausingRoundId.value === room.id
}

function roundActionLabel(room: CafeRoom): string {
  if (openingRoundId.value === room.id) return t('admin.pixelCafe.actions.openingRound')
  if (pausingRoundId.value === room.id) return t('admin.pixelCafe.actions.pausingRound')
  switch (roundStatus(room)) {
    case 'open': return t('admin.pixelCafe.actions.pauseRound')
    case 'awaiting_account': return t('admin.pixelCafe.actions.awaitingAccount')
    case 'activating': return t('admin.pixelCafe.actions.activating')
    case 'active': return t('admin.pixelCafe.actions.active')
    case 'refunding': return t('admin.pixelCafe.actions.refunding')
    default: return t('admin.pixelCafe.actions.openRound')
  }
}

function handleRoundAction(room: CafeRoom) {
  if (roundActionBusy(room)) return
  if (roundStatus(room) === 'open') {
    void pauseRound(room)
    return
  }
  if (roundStatus(room) === '') void openRound(room)
}

async function submitBulkCreate() {
  if (!canSaveBulk.value) {
    appStore.showError(t('admin.pixelCafe.bulk.noneSelected'))
    return
  }
  bulkSaving.value = true
  bulkResult.value = null
  try {
    const response = await adminAPI.cafeRooms.bulkCreate({ ...bulkForm })
    bulkResult.value = response.data
    appStore.showSuccess(t('admin.pixelCafe.success.bulkCreated', {
      created: response.data.created.length,
      failed: response.data.failed.length,
    }))
    await loadRooms()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.bulk')))
  } finally {
    bulkSaving.value = false
  }
}

onMounted(() => {
  void Promise.all([loadRooms(), loadPendingRounds()])
})

async function loadPendingRounds() { pendingLoading.value = true; try { const response = await adminAPI.cafeRooms.listPendingRounds({ page: 1, page_size: 20, search: pendingSearch.value.trim() || undefined }); pendingRounds.value = response.data.items } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.pending'))) } finally { pendingLoading.value = false } }
function openAssignDialog(round: CafePendingRound) { assigningRound.value = round; selectedAccountID.value = 0; accountSearch.value = ''; roundAccountOptions.value = []; void loadRoundAccountOptions() }
async function loadRoundAccountOptions() { if (!assigningRound.value) return; accountLoading.value = true; try { const response = await adminAPI.cafeRooms.listRoundAccountOptions(assigningRound.value.id, { page: 1, page_size: 30, search: accountSearch.value.trim() || undefined }); roundAccountOptions.value = response.data.items } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.accounts'))) } finally { accountLoading.value = false } }
async function assignAccount() { if (!assigningRound.value || !selectedAccountID.value) return; assigning.value = true; try { await adminAPI.cafeRooms.assignRoundAccount(assigningRound.value.id, selectedAccountID.value); assigningRound.value = null; appStore.showSuccess(t('admin.pixelCafe.success.accountAssigned')); await Promise.all([loadPendingRounds(), loadRooms()]) } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.pixelCafe.errors.assign'))) } finally { assigning.value = false } }

onUnmounted(() => {
  if (searchTimer) window.clearTimeout(searchTimer)
})
</script>

<style scoped>
.field {
  @apply block space-y-1;
}

.field-hint {
  @apply mt-1 block text-xs;
}

.status-badge {
  @apply inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-medium;
}

.status-badge-muted {
  @apply border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300;
}

.status-badge-success {
  @apply border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-900/20 dark:text-emerald-300;
}

.status-badge-warning {
  @apply border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-300;
}

.status-badge-danger {
  @apply border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-900/20 dark:text-red-300;
}
</style>
