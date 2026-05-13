<template>
  <AppLayout>
    <div class="purchase-pricing-page -m-4 min-h-[calc(100vh-4rem)] md:-m-[1.35rem] lg:-m-[1.6rem]">
      <div class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-10 px-4 py-8 sm:px-6 lg:px-8">
        <div v-if="loading" class="flex min-h-[28rem] items-center justify-center">
          <div class="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent"></div>
        </div>
        <template v-else>
          <template v-if="paymentPhase === 'paying'">
            <div class="pricing-card mx-auto w-full max-w-3xl rounded-2xl p-4 sm:p-6">
              <PaymentStatusPanel
                :order-id="paymentState.orderId"
                :qr-code="paymentState.qrCode"
                :expires-at="paymentState.expiresAt"
                :payment-type="paymentState.paymentType"
                :pay-url="paymentState.payUrl"
                :order-type="paymentState.orderType"
                :currency="paymentState.currency || selectedCurrency"
                @done="onPaymentDone"
                @success="onPaymentSuccess"
                @settled="onPaymentSettled"
              />
            </div>
          </template>
          <template v-else>
            <header class="mx-auto flex max-w-3xl flex-col items-center gap-5 text-center">
              <div>
                <h1 class="pricing-title text-3xl font-black tracking-normal sm:text-4xl">{{ pt('title') }}</h1>
                <p class="pricing-subtitle mt-3 text-sm leading-6 sm:text-base">{{ pt('subtitle') }}</p>
              </div>
              <div class="pricing-current-chip inline-flex max-w-full flex-wrap items-center justify-center gap-2 rounded-lg px-4 py-3 text-sm">
                <Icon name="infoCircle" size="sm" class="pricing-muted-icon" />
                <span>{{ pt('currentPlan') }}:</span>
                <strong class="pricing-strong">{{ currentPlanName }}</strong>
                <span class="pricing-separator">&middot;</span>
                <span>{{ pt('balance') }}:</span>
                <strong class="pricing-strong">{{ currentBalanceText }}</strong>
              </div>
            </header>

            <div v-if="errorMessage" class="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-400/30 dark:bg-amber-500/10 dark:text-amber-100">
              <div class="flex gap-3">
                <Icon name="exclamationTriangle" size="sm" class="mt-0.5 shrink-0 text-amber-500 dark:text-amber-300" />
                <div>
                  <p class="font-semibold">{{ errorMessage }}</p>
                  <p v-if="errorHintMessage" class="mt-1 text-amber-700/80 dark:text-amber-100/75">{{ errorHintMessage }}</p>
                </div>
              </div>
            </div>

            <section v-if="!checkout.balance_disabled" class="space-y-6">
              <div class="flex flex-wrap items-center gap-3">
                <Icon name="creditCard" size="lg" class="text-sky-400" />
                <h2 class="pricing-section-title text-xl font-black tracking-normal">{{ pt('flexibleCredit') }}</h2>
                <span class="pricing-section-tag rounded-md px-2.5 py-1 text-xs font-medium">{{ pt('creditTag') }}</span>
              </div>

              <div v-if="enabledMethods.length === 0" class="pricing-card rounded-3xl py-16 text-center">
                <p class="pricing-muted">{{ t('payment.notAvailable') }}</p>
              </div>
              <div v-else class="pricing-card rounded-3xl p-5 sm:p-8">
                <div class="grid gap-8 lg:grid-cols-[minmax(0,1.45fr)_minmax(320px,0.95fr)]">
                  <div class="space-y-6">
                    <div>
                      <p class="pricing-strong text-sm font-bold">{{ pt('rechargeStep') }}</p>
                      <div class="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-3">
                        <button
                          v-for="preset in rechargeAmountOptions"
                          :key="preset"
                          type="button"
                          class="pricing-preset relative rounded-xl border px-4 py-4 text-center transition"
                          :class="[
                            validAmount === preset
                              ? 'pricing-preset--selected'
                              : 'pricing-preset--idle',
                            rechargePresetAvailable(preset) ? '' : 'cursor-not-allowed opacity-40'
                          ]"
                          :disabled="!rechargePresetAvailable(preset)"
                          @click="selectRechargeAmount(preset)"
                        >
                          <span v-if="showRecommendedRechargeBadge && preset === recommendedRechargeAmount" class="absolute -top-2 left-1/2 -translate-x-1/2 rounded-full bg-orange-500 px-2 py-0.5 text-[10px] font-bold text-white">
                            {{ pt('newUserDeal') }}
                          </span>
                          <span class="block text-2xl font-black">{{ formatSelectedPaymentAmount(preset) }}</span>
                          <span class="pricing-caption mt-2 block text-xs">{{ formatCreditAmount(preset) }}</span>
                        </button>
                      </div>
                    </div>

                    <div class="pricing-subpanel rounded-xl p-4">
                      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                        <div>
                          <p class="pricing-strong text-sm font-semibold">{{ pt('customUnits') }}</p>
                          <p class="pricing-caption mt-1 text-xs">{{ pt('unitHint', { amount: formatSelectedPaymentAmount(rechargeUnitAmount) }) }}</p>
                        </div>
                        <div class="pricing-stepper inline-flex h-10 w-32 shrink-0 items-center overflow-hidden rounded-lg">
                          <button type="button" class="pricing-stepper-button flex h-full w-10 items-center justify-center text-lg transition disabled:opacity-40" :disabled="rechargeUnits <= 1" @click="decreaseRechargeUnits">-</button>
                          <span class="pricing-stepper-value flex-1 text-center text-sm font-bold">{{ rechargeUnits }}</span>
                          <button type="button" class="pricing-stepper-button flex h-full w-10 items-center justify-center transition" @click="increaseRechargeUnits">
                            <Icon name="plus" size="sm" />
                          </button>
                        </div>
                      </div>
                      <label class="sr-only" for="custom-recharge-amount">{{ t('payment.customAmount') }}</label>
                      <input
                        id="custom-recharge-amount"
                        class="pricing-input mt-4 w-full rounded-lg border px-3 py-2 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 dark:focus:border-blue-400"
                        type="number"
                        inputmode="decimal"
                        min="0"
                        :value="amount ?? ''"
                        :placeholder="t('payment.enterAmount')"
                        @input="setCustomAmountFromEvent"
                      >
                      <p v-if="amountError" class="mt-2 text-xs text-amber-600 dark:text-amber-300">{{ amountError }}</p>
                    </div>

                    <div v-if="methodOptions.length > 1" class="space-y-3">
                      <p class="pricing-strong text-sm font-bold">{{ t('payment.paymentMethod') }}</p>
                      <div class="grid gap-2 sm:grid-cols-2">
                        <button
                          v-for="method in methodOptions"
                          :key="method.type"
                          type="button"
                          class="pricing-method-option rounded-lg border px-3 py-2 text-left text-sm transition"
                          :class="[
                            selectedMethod === method.type
                              ? 'pricing-method-option--selected'
                              : 'pricing-method-option--idle',
                            method.available ? '' : 'cursor-not-allowed opacity-45'
                          ]"
                          :disabled="!method.available"
                          @click="selectedMethod = method.type"
                        >
                          <span class="block font-semibold">{{ paymentMethodLabel(method.type) }}</span>
                          <span v-if="method.fee_rate > 0" class="pricing-caption mt-1 block text-xs">{{ t('payment.fee') }} {{ method.fee_rate }}%</span>
                        </button>
                      </div>
                    </div>
                  </div>

                  <aside class="pricing-summary rounded-2xl p-6">
                    <h3 class="pricing-strong text-sm font-black">{{ pt('orderSummary') }}</h3>
                    <div class="mt-6 space-y-4 text-sm">
                      <div class="flex items-center justify-between gap-4">
                        <span class="pricing-muted">{{ pt('rechargeAmount') }}</span>
                        <span class="pricing-strong font-semibold">{{ formatSelectedPaymentAmount(validAmount) }}</span>
                      </div>
                      <div v-if="showRechargeUnitsSummary" class="flex items-center justify-between gap-4">
                        <span class="pricing-muted">{{ pt('purchaseUnits') }}</span>
                        <span class="pricing-strong font-semibold">{{ rechargeUnits }} {{ pt('units') }}</span>
                      </div>
                      <div class="flex items-center justify-between gap-4">
                        <span class="pricing-muted">{{ t('payment.creditedBalance') }}</span>
                        <span class="pricing-strong font-semibold">{{ formatCreditAmount(validAmount) }}</span>
                      </div>
                      <div v-if="selectedMethodLabel" class="flex items-center justify-between gap-4">
                        <span class="pricing-muted">{{ t('payment.paymentMethod') }}</span>
                        <span class="pricing-strong font-semibold">{{ selectedMethodLabel }}</span>
                      </div>
                      <div v-if="feeRate > 0" class="flex items-center justify-between gap-4">
                        <span class="pricing-muted">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                        <span class="pricing-strong font-semibold">{{ formatSelectedPaymentAmount(feeAmount) }}</span>
                      </div>
                      <div class="pricing-divider border-t border-dashed pt-6">
                        <div class="flex items-end justify-between gap-4">
                          <span class="pricing-muted">{{ pt('totalPayable') }}</span>
                          <span class="min-w-0 break-words text-right text-3xl font-black text-blue-600 tabular-nums dark:text-blue-400 sm:text-4xl">{{ formatSelectedPaymentAmount(totalAmount) }}</span>
                        </div>
                      </div>
                    </div>
                    <button :class="['btn mt-7 inline-flex w-full items-center justify-center gap-2 py-3 text-base font-bold', paymentButtonClass]" :disabled="!canSubmit || submitting" @click="handleSubmitRecharge">
                      <span v-if="submitting" class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                      <Icon v-else name="shield" size="sm" />
                      {{ submitting ? t('common.processing') : pt('paySecurely') }}
                    </button>
                  </aside>
                </div>
              </div>
            </section>

            <section class="pricing-section-divider space-y-6 border-t pt-10">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="flex flex-wrap items-center gap-3">
                  <Icon name="creditCard" size="lg" class="text-sky-400" />
                  <h2 class="pricing-section-title text-xl font-black tracking-normal">{{ pt('plansTitle') }}</h2>
                  <span class="pricing-section-tag rounded-md px-2.5 py-1 text-xs font-medium">{{ pt('plansTag') }}</span>
                </div>
                <button
                  type="button"
                  class="pricing-refresh-button inline-flex h-9 w-9 items-center justify-center rounded-lg transition disabled:cursor-wait disabled:opacity-60"
                  :disabled="checkoutRefreshing"
                  :title="t('common.refresh')"
                  @click="refreshCheckoutInfo"
                >
                  <Icon name="refresh" size="sm" :class="checkoutRefreshing ? 'animate-spin' : ''" />
                </button>
              </div>

              <div v-if="checkout.plans.length === 0" class="pricing-card rounded-3xl py-16 text-center">
                <Icon name="gift" size="xl" class="pricing-empty-icon mx-auto mb-3" />
                <p class="pricing-muted">{{ t('payment.noPlans') }}</p>
              </div>
              <div v-else class="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
                <article
                  v-for="plan in checkout.plans"
                  :key="plan.id"
                  class="pricing-plan-card relative flex min-h-[26rem] flex-col rounded-3xl border p-8 transition"
                  :class="[
                    selectedPlan?.id === plan.id
                      ? 'pricing-plan-card--selected'
                      : isRecommendedPlan(plan)
                        ? 'pricing-plan-card--recommended'
                        : 'pricing-plan-card--idle'
                  ]"
                >
                  <span v-if="isRecommendedPlan(plan)" class="absolute left-1/2 top-0 -translate-x-1/2 -translate-y-1/2 rounded-full bg-indigo-500 px-4 py-1 text-xs font-bold text-white">
                    {{ pt('recommended') }}
                  </span>
                  <div class="text-center">
                    <div class="flex justify-center">
                      <span :class="['rounded-full px-2 py-0.5 text-[10px] font-semibold', platformBadgeLightClass(plan.group_platform || '')]">
                        {{ platformLabel(plan.group_platform || '') }}
                      </span>
                    </div>
                    <h3 class="pricing-strong mt-5 text-xl font-black">{{ plan.name }}</h3>
                    <p class="pricing-caption mt-2 min-h-[2.5rem] text-sm leading-5">{{ plan.description || pt('planDefaultDesc') }}</p>
                    <div class="mt-6 flex min-w-0 flex-wrap items-end justify-center gap-2">
                      <span v-if="plan.original_price" class="pricing-strike pb-2 text-sm line-through">{{ formatSelectedPaymentAmount(plan.original_price) }}</span>
                      <span class="pricing-plan-price break-words text-center text-4xl font-black tabular-nums sm:text-5xl">{{ formatSelectedPaymentAmount(plan.price) }}</span>
                      <span class="pricing-muted pb-2 text-sm">/{{ planValiditySuffixFor(plan) }}</span>
                    </div>
                  </div>

                  <div class="pricing-plan-divider my-7 h-px"></div>

                  <ul class="pricing-feature-list flex-1 space-y-4 text-sm">
                    <li v-for="feature in planFeatureList(plan)" :key="feature" class="flex gap-3">
                      <Icon name="check" size="sm" class="mt-0.5 shrink-0 text-blue-500 dark:text-blue-300" />
                      <span>{{ feature }}</span>
                    </li>
                  </ul>

                  <button
                    type="button"
                    class="btn mt-8 inline-flex w-full items-center justify-center gap-2 py-3 text-sm font-bold"
                    :class="isRecommendedPlan(plan) || selectedPlan?.id === plan.id ? 'btn-primary' : 'btn-secondary'"
                    @click="selectPlan(plan)"
                  >
                    <Icon name="arrowRight" size="sm" />
                    {{ isPlanRenewal(plan) ? t('payment.renewNow') : pt('subscribeCta') }}
                  </button>
                </article>
              </div>

              <div v-if="selectedPlan" class="pricing-confirm-panel rounded-3xl p-5 sm:p-6">
                <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(280px,0.8fr)]">
                  <div>
                    <p class="text-sm font-bold text-blue-700 dark:text-blue-100">{{ pt('confirmTitle') }}</p>
                    <h3 class="pricing-strong mt-2 text-2xl font-black">{{ selectedPlan.name }}</h3>
                    <p class="pricing-muted mt-2 text-sm">{{ selectedPlan.description || pt('planDefaultDesc') }}</p>

                    <div v-if="subMethodOptions.length > 1" class="mt-5 space-y-3">
                      <p class="pricing-strong text-sm font-bold">{{ t('payment.paymentMethod') }}</p>
                      <div class="grid gap-2 sm:grid-cols-2">
                        <button
                          v-for="method in subMethodOptions"
                          :key="method.type"
                          type="button"
                          class="pricing-method-option rounded-lg border px-3 py-2 text-left text-sm transition"
                          :class="[
                            selectedMethod === method.type
                              ? 'pricing-method-option--selected'
                              : 'pricing-method-option--idle',
                            method.available ? '' : 'cursor-not-allowed opacity-45'
                          ]"
                          :disabled="!method.available"
                          @click="selectedMethod = method.type"
                        >
                          <span class="block font-semibold">{{ paymentMethodLabel(method.type) }}</span>
                          <span v-if="method.fee_rate > 0" class="pricing-caption mt-1 block text-xs">{{ t('payment.fee') }} {{ method.fee_rate }}%</span>
                        </button>
                      </div>
                    </div>
                  </div>

                  <div class="pricing-summary rounded-2xl p-5">
                    <div class="space-y-3 text-sm">
                      <div class="flex justify-between gap-4">
                        <span class="pricing-muted">{{ pt('selectedPlan') }}</span>
                        <span class="pricing-strong font-semibold">{{ selectedPlan.name }}</span>
                      </div>
                      <div class="flex justify-between gap-4">
                        <span class="pricing-muted">{{ pt('subtotal') }}</span>
                        <span class="pricing-strong font-semibold">{{ formatSelectedPaymentAmount(selectedPlan.price) }}</span>
                      </div>
                      <div v-if="feeRate > 0 && selectedPlan.price > 0" class="flex justify-between gap-4">
                        <span class="pricing-muted">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                        <span class="pricing-strong font-semibold">{{ formatSelectedPaymentAmount(subFeeAmount) }}</span>
                      </div>
                      <div class="pricing-divider border-t border-dashed pt-5">
                        <div class="flex items-end justify-between gap-4">
                          <span class="pricing-muted">{{ pt('totalPayable') }}</span>
                          <span class="min-w-0 break-words text-right text-3xl font-black text-blue-600 tabular-nums dark:text-blue-400">{{ formatSelectedPaymentAmount(subTotalAmount) }}</span>
                        </div>
                      </div>
                    </div>
                    <div class="mt-6 grid gap-2 sm:grid-cols-2">
                      <button :class="['btn inline-flex items-center justify-center gap-2 py-3 text-sm font-bold', paymentButtonClass]" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                        <span v-if="submitting" class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                        <Icon v-else name="shield" size="sm" />
                        {{ submitting ? t('common.processing') : pt('subscribeCta') }}
                      </button>
                      <button class="btn btn-secondary py-3 text-sm font-bold" type="button" @click="selectedPlan = null">{{ t('common.cancel') }}</button>
                    </div>
                  </div>
                </div>
              </div>

              <div v-if="activeSubscriptions.length > 0">
                <p class="pricing-muted mb-3 text-sm font-bold">{{ t('payment.activeSubscription') }}</p>
                <div class="grid gap-2 md:grid-cols-2">
                  <div v-for="sub in activeSubscriptions" :key="sub.id"
                    class="pricing-subscription-card flex items-center gap-3 rounded-xl px-3 py-2">
                    <div :class="['h-6 w-1 shrink-0 rounded-full', platformAccentBarClass(sub.group?.platform || '')]" />
                    <div class="min-w-0 flex-1">
                      <div class="flex items-center gap-1.5">
                        <span class="pricing-strong truncate text-xs font-semibold">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                        <span :class="['shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                      </div>
                      <div class="pricing-caption flex flex-wrap gap-x-3 text-[11px]">
                        <span>{{ t('payment.planCard.rate') }}: ×{{ sub.group?.rate_multiplier ?? 1 }}</span>
                        <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                        <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                        <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                      </div>
                    </div>
                    <span class="badge badge-success shrink-0 text-[10px]">{{ t('userSubscriptions.status.active') }}</span>
                  </div>
                </div>
              </div>
            </section>

            <section class="mx-auto w-full max-w-3xl space-y-4 pt-4">
              <h2 class="pricing-section-title text-center text-xl font-black">{{ pt('faqTitle') }}</h2>
              <div class="pricing-faq-list">
                <div v-for="(item, index) in faqItems" :key="item.title">
                  <button class="pricing-faq-button flex w-full items-center justify-between gap-4 py-4 text-left text-sm font-bold" type="button" @click="openFaqIndex = openFaqIndex === index ? null : index">
                    <span>{{ item.title }}</span>
                    <Icon name="chevronDown" size="sm" class="pricing-muted-icon shrink-0 transition" :class="{ 'rotate-180': openFaqIndex === index }" />
                  </button>
                  <p v-if="openFaqIndex === index" class="pricing-muted pb-4 text-sm leading-6">{{ item.body }}</p>
                </div>
              </div>
            </section>

            <section v-if="checkout.help_text || checkout.help_image_url" class="pricing-card rounded-3xl p-5">
              <div class="flex flex-col items-center gap-3">
                <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
                  class="h-40 max-w-full cursor-pointer rounded-lg object-contain transition-opacity hover:opacity-80"
                  @click="previewImage = checkout.help_image_url" />
                <p v-if="checkout.help_text" class="pricing-muted text-center text-sm">{{ checkout.help_text }}</p>
              </div>
            </section>
          </template>
        </template>
      </div>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="closeRenewalModal">
          <div class="relative w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-dark-700 dark:bg-dark-900">
            <!-- Close button -->
            <button class="absolute right-4 top-4 rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200" @click="closeRenewalModal">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
            <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-4">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="max-h-[85vh] max-w-[90vw] rounded-xl object-contain shadow-2xl" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  type VisiblePaymentMethod,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import { platformAccentBarClass, platformBadgeLightClass, platformLabel } from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

type PricingLocale = 'en' | 'zh'
type PricingMessage = string | { [key: string]: PricingMessage }

const pricingCatalog = {
  zh: {
    title: '定价方案',
    subtitle: '按需充值或选择超值订阅，以更低成本调用全网大模型。',
    currentPlan: '当前套餐',
    balance: '灵活额度',
    freePlan: 'Free Plan',
    flexibleCredit: '灵活额度',
    creditTag: '额度用完前，永久有效',
    rechargeStep: '第一步：选择或输入充值金额',
    newUserDeal: '新人专享',
    customUnits: '自定义购买份数',
    unitHint: '每份按 {amount} 递增',
    orderSummary: '订单摘要',
    rechargeAmount: '充值额度',
    purchaseUnits: '购买份数',
    units: '份',
    totalPayable: '总计应付',
    paySecurely: '立即安全支付',
    plansTitle: '套餐订阅',
    plansTag: '周期额度每周期重置，套餐内有效，过期不补',
    recommended: '推荐选择',
    planDefaultDesc: '适合稳定调用全网大模型的开发者。',
    subscribeCta: '立即订阅',
    confirmTitle: '确认订阅套餐',
    selectedPlan: '已选套餐',
    subtotal: '套餐金额',
    faqTitle: '常见问题',
    feature: {
      weeklyQuota: '周额度 {amount}',
      monthlyQuota: '月额度 {amount}',
      dailyQuota: '日额度 {amount}',
      discountRate: '相当于全站 API 计费额外尊享 {rate} 折',
      modelScopes: '覆盖 {count} 个模型范围',
      unlimitedQuota: '不限制套餐内额度',
    },
    faq: {
      quota: {
        title: '额度与计费规则',
        body: '灵活额度按实际调用消耗；订阅套餐按配置的周期额度和倍率计费，具体以当前站点配置为准。',
      },
      balance: {
        title: '灵活额度说明',
        body: '灵活额度充值后进入账户余额，在额度用完前持续有效，可用于未被套餐覆盖的调用。',
      },
      upgrade: {
        title: '如何升级套餐？',
        body: '选择更高档套餐并完成支付后，系统会按当前订阅规则刷新可用额度和有效期。',
      },
      recovery: {
        title: '额度恢复机制',
        body: '订阅额度按套餐周期自动重置；灵活额度不会周期清零，只随调用扣减。',
      },
      subscription: {
        title: '套餐变更说明',
        body: '同一订阅分组再次购买通常视为续费或延长，具体生效方式由后台套餐配置决定。',
      },
      balanceSubscription: {
        title: '订阅额度与灵活额度',
        body: '优先使用订阅套餐覆盖的额度；超出或未覆盖部分可继续使用灵活额度支付。',
      },
    },
  },
  en: {
    title: 'Pricing',
    subtitle: 'Top up as needed or choose a subscription to call all available models at a lower cost.',
    currentPlan: 'Current plan',
    balance: 'Flexible credit',
    freePlan: 'Free Plan',
    flexibleCredit: 'Flexible Credit',
    creditTag: 'Valid until used up',
    rechargeStep: 'Step 1: choose or enter a recharge amount',
    newUserDeal: 'New user',
    customUnits: 'Custom purchase units',
    unitHint: 'Each unit increases by {amount}',
    orderSummary: 'Order Summary',
    rechargeAmount: 'Recharge credit',
    purchaseUnits: 'Units',
    units: 'units',
    totalPayable: 'Total due',
    paySecurely: 'Pay Securely',
    plansTitle: 'Subscriptions',
    plansTag: 'Quota resets by cycle and does not roll over after expiry',
    recommended: 'Recommended',
    planDefaultDesc: 'For developers who need stable access to all available models.',
    subscribeCta: 'Subscribe Now',
    confirmTitle: 'Confirm Subscription',
    selectedPlan: 'Selected plan',
    subtotal: 'Subtotal',
    faqTitle: 'FAQ',
    feature: {
      weeklyQuota: 'Weekly quota {amount}',
      monthlyQuota: 'Monthly quota {amount}',
      dailyQuota: 'Daily quota {amount}',
      discountRate: 'Effective API billing rate: {rate}x',
      modelScopes: 'Covers {count} model scopes',
      unlimitedQuota: 'No quota limit inside this plan',
    },
    faq: {
      quota: {
        title: 'Quota and billing rules',
        body: 'Flexible credit is consumed by actual usage. Subscription plans follow the configured cycle quota and rate multiplier.',
      },
      balance: {
        title: 'Flexible credit',
        body: 'Top-up credit is added to your account balance and remains valid until it is consumed.',
      },
      upgrade: {
        title: 'How do I upgrade?',
        body: 'Choose a higher plan and complete payment. The system refreshes quota and validity according to the plan rules.',
      },
      recovery: {
        title: 'Quota recovery',
        body: 'Subscription quota resets by plan cycle. Flexible credit does not reset periodically and is only reduced by usage.',
      },
      subscription: {
        title: 'Plan changes',
        body: 'Buying a plan in the same subscription group usually renews or extends it, depending on the backend plan settings.',
      },
      balanceSubscription: {
        title: 'Subscription quota and credit',
        body: 'Subscription quota is used for covered requests first. Uncovered or over-quota usage can continue with flexible credit.',
      },
    },
  },
} satisfies Record<PricingLocale, PricingMessage>

function interpolatePricingText(text: string, params?: Record<string, string | number>): string {
  if (!params) return text
  return Object.entries(params).reduce((result, [key, value]) => result.split(`{${key}}`).join(String(value)), text)
}

function readPricingFallback(path: string): string {
  const locale = String(localeCode.value || '').toLowerCase().startsWith('zh') ? 'zh' : 'en'
  const value = path.split('.').reduce<PricingMessage | undefined>((current, part) => {
    if (!current || typeof current === 'string') return undefined
    return current[part]
  }, pricingCatalog[locale])
  return typeof value === 'string' ? value : path
}

function pt(path: string, params?: Record<string, string | number>): string {
  const key = `payment.pricing.${path}`
  const translated = t(key, params ?? {})
  if (translated && translated !== key && !translated.startsWith('payment.pricing.')) {
    return translated
  }
  return interpolatePricingText(readPricingFallback(path), params)
}

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

const loading = ref(true)
const submitting = ref(false)
const checkoutRefreshing = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const amount = ref<number | null>(null)
const selectedMethod = ref('')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const previewImage = ref('')
const openFaqIndex = ref<number | null>(null)

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface PaymentMethodOption {
  type: string
  fee_rate: number
  available: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_fee_rate: 0, help_text: '', help_image_url: '', stripe_publishable_key: '',
})

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const enabledMethods = computed<VisiblePaymentMethod[]>(() => Object.keys(visibleMethods.value) as VisiblePaymentMethod[])
const validAmount = computed(() => amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return multiplier > 0 ? multiplier : 1
})
const rechargeUnitAmount = 1
const rechargeAmountPresets = [5, 50, 200, 300, 500, 1000]
const recommendedRechargeAmount = 5
const showRecommendedRechargeBadge = false

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  if (ml.single_min > 0 && amt < ml.single_min) return false
  if (ml.single_max > 0 && amt > ml.single_max) return false
  return true
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => limit.single_min))
})
const globalMaxAmount = computed(() => {
  const limits = Object.values(visibleMethods.value)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => limit.single_max))
})

// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() => visibleMethods.value[selectedMethod.value])
const selectedCurrency = computed(() => normalizePaymentCurrency(selectedLimit.value?.currency))
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function formatSelectedPaymentAmount(value: number): string {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value)
}

function formatCreditAmount(value: number): string {
  const credited = Math.round((value * balanceRechargeMultiplier.value) * 100) / 100
  return `$${credited.toFixed(2)} USD`
}

const currentBalanceText = computed(() => `$${(user.value?.balance ?? 0).toFixed(2)}`)
const currentPlanName = computed(() => {
  const subscription = activeSubscriptions.value[0]
  return subscription?.group?.name || pt('freePlan')
})

const rechargeAmountOptions = computed(() =>
  rechargeAmountPresets.filter((preset) => {
    if (globalMinAmount.value > 0 && preset < globalMinAmount.value) return false
    if (globalMaxAmount.value > 0 && preset > globalMaxAmount.value) return false
    return enabledMethods.value.some((method) => amountFitsMethod(preset, method))
  })
)

const rechargeUnits = computed(() => Math.max(1, Math.round((validAmount.value || rechargeUnitAmount) / rechargeUnitAmount)))
const showRechargeUnitsSummary = computed(() =>
  validAmount.value >= rechargeUnitAmount && validAmount.value % rechargeUnitAmount === 0
)

const selectedMethodLabel = computed(() => selectedMethod.value ? paymentMethodLabel(selectedMethod.value) : '')

const recommendedPlanId = computed(() => {
  const plans = checkout.value.plans
  if (plans.length === 0) return 0
  if (plans.length >= 2) return plans[1].id
  return plans[0].id
})

const faqItems = computed(() => [
  {
    title: pt('faq.quota.title'),
    body: pt('faq.quota.body'),
  },
  {
    title: pt('faq.balance.title'),
    body: pt('faq.balance.body'),
  },
  {
    title: pt('faq.upgrade.title'),
    body: pt('faq.upgrade.body'),
  },
  {
    title: pt('faq.recovery.title'),
    body: pt('faq.recovery.body'),
  },
  {
    title: pt('faq.subscription.title'),
    body: pt('faq.subscription.body'),
  },
  {
    title: pt('faq.balanceSubscription.title'),
    body: pt('faq.balanceSubscription.body'),
  },
])

function paymentMethodLabel(type: string): string {
  return t(`payment.methods.${type}`)
}

function rechargePresetAvailable(preset: number): boolean {
  return enabledMethods.value.some((method) => amountFitsMethod(preset, method))
}

function selectDefaultMethodIfNeeded(force = false) {
  if (!enabledMethods.value.length) {
    selectedMethod.value = ''
    return
  }
  const currentVisibleMethod = normalizeVisibleMethod(selectedMethod.value)
  if (!force && currentVisibleMethod && enabledMethods.value.includes(currentVisibleMethod)) {
    return
  }
  const order: readonly string[] = METHOD_ORDER
  const sorted = [...enabledMethods.value].sort((a, b) => {
    const ai = order.indexOf(a)
    const bi = order.indexOf(b)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })
  selectedMethod.value = sorted[0]
}

async function loadCheckoutInfo(preservePlan = true) {
  const previousPlanId = preservePlan ? selectedPlan.value?.id : undefined
  const res = await paymentAPI.getCheckoutInfo()
  checkout.value = res.data
  selectDefaultMethodIfNeeded()
  if (previousPlanId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === previousPlanId) ?? null
  }
}

async function refreshCheckoutInfo() {
  if (checkoutRefreshing.value) return
  checkoutRefreshing.value = true
  try {
    await loadCheckoutInfo()
    await subscriptionStore.fetchActiveSubscriptions(true)
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    checkoutRefreshing.value = false
  }
}

function selectRechargeAmount(preset: number) {
  if (!rechargePresetAvailable(preset)) return
  amount.value = preset
}

function setRechargeUnits(units: number) {
  const nextUnits = Math.max(1, units)
  amount.value = nextUnits * rechargeUnitAmount
}

function decreaseRechargeUnits() {
  setRechargeUnits(rechargeUnits.value - 1)
}

function increaseRechargeUnits() {
  setRechargeUnits(rechargeUnits.value + 1)
}

function setCustomAmountFromEvent(event: Event) {
  const target = event.target as HTMLInputElement | null
  const value = Number.parseFloat(target?.value || '')
  amount.value = Number.isFinite(value) && value > 0 ? value : null
}

function planValiditySuffixFor(plan: SubscriptionPlan): string {
  const unit = plan.validity_unit || 'day'
  if (unit === 'month') return t('payment.perMonth')
  if (unit === 'year') return t('payment.perYear')
  return `${plan.validity_days}${t('payment.days')}`
}

function isRecommendedPlan(plan: SubscriptionPlan): boolean {
  return plan.id === recommendedPlanId.value
}

function isPlanRenewal(plan: SubscriptionPlan): boolean {
  return activeSubscriptions.value.some((subscription) => subscription.group_id === plan.group_id)
}

function planFeatureList(plan: SubscriptionPlan): string[] {
  const features: string[] = []
  if (plan.weekly_limit_usd != null) features.push(pt('feature.weeklyQuota', { amount: `$${plan.weekly_limit_usd}` }))
  if (plan.monthly_limit_usd != null) features.push(pt('feature.monthlyQuota', { amount: `$${plan.monthly_limit_usd}` }))
  if (plan.daily_limit_usd != null) features.push(pt('feature.dailyQuota', { amount: `$${plan.daily_limit_usd}` }))
  if (plan.rate_multiplier != null && plan.rate_multiplier !== 1) {
    features.push(pt('feature.discountRate', { rate: plan.rate_multiplier }))
  }
  if (plan.supported_model_scopes?.length) {
    features.push(pt('feature.modelScopes', { count: plan.supported_model_scopes.length }))
  }
  features.push(...(plan.features || []).filter(Boolean))
  if (features.length === 0) {
    features.push(pt('feature.unlimitedQuota'))
  }
  return features.slice(0, 6)
}

const methodOptions = computed<PaymentMethodOption[]>(() =>
  enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)
const feeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const totalAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + feeAmount.value) * 100) / 100
    : validAmount.value
)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!enabledMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedLimit.value
  if (ml) {
    if (ml.single_min > 0 && validAmount.value < ml.single_min) return t('payment.amountTooLow', { min: formatSelectedPaymentAmount(ml.single_min) })
    if (ml.single_max > 0 && validAmount.value > ml.single_max) return t('payment.amountTooHigh', { max: formatSelectedPaymentAmount(ml.single_max) })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && amountFitsMethod(validAmount.value, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Subscription-specific: method options based on plan price
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const planPrice = selectedPlan.value?.price ?? 0
  return enabledMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(planPrice, type),
    }
  })
})

const subFeeAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return 0
  return Math.ceil(((price * feeRate.value) / 100) * 100) / 100
})

const subTotalAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (feeRate.value <= 0 || price <= 0) return price
  return Math.round((price + subFeeAmount.value) * 100) / 100
})

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && amountFitsMethod(selectedPlan.value.price, selectedMethod.value)
    && selectedLimit.value?.available !== false
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedMethod.value] as const, ([amt, method]) => {
  if (amt <= 0 || amountFitsMethod(amt, method)) return
  const available = enabledMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedMethod.value = available
})

watch(() => selectedPlan.value?.price ?? 0, (planPrice) => {
  if (planPrice <= 0 || amountFitsMethod(planPrice, selectedMethod.value)) return
  const available = enabledMethods.value.find((method) =>
    visibleMethods.value[method]?.available !== false && amountFitsMethod(planPrice, method)
  )
  if (available) selectedMethod.value = available
})

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = selectedMethod.value
  if (!m) return 'btn-primary'
  if (m.includes('alipay')) return 'btn-alipay'
  if (m.includes('wxpay')) return 'btn-wxpay'
  if (m === 'stripe') return 'btn-stripe'
  if (m === 'airwallex') return 'btn-airwallex'
  return 'btn-primary'
})

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  await createOrder(validAmount.value, 'balance')
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id)
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const requestType = normalizeVisibleMethod(options.paymentType || selectedMethod.value) || options.paymentType || selectedMethod.value
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentType: requestType,
      orderType,
      planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || selectedMethod.value) || selectedMethod.value,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  orderType: OrderType
  planId?: number
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  selectedMethod.value = resume.paymentType
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
    })
  }
}

onMounted(async () => {
  try {
    await loadCheckoutInfo(false)
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
        if (restoredMethod) {
          selectedMethod.value = restoredMethod
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (!checkout.value.balance_disabled && amount.value == null && !hasWechatResumeQuery(route.query)) {
      amount.value = rechargeAmountOptions.value.includes(1000)
        ? 1000
        : (rechargeAmountOptions.value[0] ?? rechargeUnitAmount)
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>

<style scoped>
.purchase-pricing-page {
  position: relative;
  z-index: 2;
  isolation: isolate;
  overflow: hidden;
  min-width: 0;
  background:
    radial-gradient(circle at 50% 0%, rgba(59, 130, 246, 0.1), transparent 30rem),
    linear-gradient(180deg, rgba(248, 250, 252, 0.96), rgba(239, 246, 255, 0.86) 46%, rgba(248, 250, 252, 0.94));
  color: #0f172a;
}

:global(.dark) .purchase-pricing-page {
  background:
    radial-gradient(circle at 50% 0%, rgba(59, 130, 246, 0.12), transparent 30rem),
    #0c0d10;
  color: #f8fafc;
}

.purchase-pricing-page::before {
  position: absolute;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  content: '';
  opacity: 0.34;
  background-image:
    linear-gradient(rgba(37, 99, 235, 0.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(37, 99, 235, 0.07) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: linear-gradient(to bottom, black, transparent 78%);
}

:global(.dark) .purchase-pricing-page::before {
  opacity: 0.22;
  background-image:
    linear-gradient(rgba(148, 163, 184, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(148, 163, 184, 0.08) 1px, transparent 1px);
}

.pricing-title,
.pricing-section-title,
.pricing-strong,
.pricing-faq-button {
  color: #0f172a;
}

:global(.dark) .pricing-title,
:global(.dark) .pricing-section-title,
:global(.dark) .pricing-strong,
:global(.dark) .pricing-faq-button {
  color: #f8fafc;
}

.pricing-subtitle,
.pricing-muted,
.pricing-caption {
  color: #475569;
}

:global(.dark) .pricing-subtitle,
:global(.dark) .pricing-muted,
:global(.dark) .pricing-caption {
  color: #94a3b8;
}

.pricing-caption {
  color: #64748b;
}

:global(.dark) .pricing-caption {
  color: #94a3b8;
}

.pricing-muted-icon {
  color: #64748b;
}

:global(.dark) .pricing-muted-icon {
  color: #94a3b8;
}

.pricing-separator,
.pricing-strike {
  color: #94a3b8;
}

:global(.dark) .pricing-separator,
:global(.dark) .pricing-strike {
  color: #64748b;
}

.pricing-current-chip,
.pricing-card,
.pricing-plan-card,
.pricing-subpanel,
.pricing-summary,
.pricing-subscription-card,
.pricing-confirm-panel {
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(255, 255, 255, 0.88);
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.07);
  backdrop-filter: blur(16px);
}

:global(.dark) .pricing-current-chip,
:global(.dark) .pricing-card,
:global(.dark) .pricing-plan-card,
:global(.dark) .pricing-subpanel,
:global(.dark) .pricing-summary,
:global(.dark) .pricing-subscription-card,
:global(.dark) .pricing-confirm-panel {
  border-color: rgba(51, 65, 85, 0.9);
  background: rgba(17, 24, 39, 0.84);
  box-shadow: 0 18px 48px rgba(0, 0, 0, 0.22);
}

.pricing-current-chip {
  color: #475569;
}

:global(.dark) .pricing-current-chip {
  color: #cbd5e1;
}

.pricing-section-tag {
  background: rgba(219, 234, 254, 0.86);
  color: #1d4ed8;
}

:global(.dark) .pricing-section-tag {
  background: rgba(30, 41, 59, 0.94);
  color: #cbd5e1;
}

.pricing-refresh-button {
  border: 1px solid rgba(148, 163, 184, 0.28);
  background: rgba(255, 255, 255, 0.72);
  color: #475569;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.08);
}

.pricing-refresh-button:hover:not(:disabled) {
  border-color: rgba(59, 130, 246, 0.36);
  color: #2563eb;
}

:global(.dark) .pricing-refresh-button {
  border-color: rgba(148, 163, 184, 0.22);
  background: rgba(15, 23, 42, 0.72);
  color: #cbd5e1;
  box-shadow: none;
}

:global(.dark) .pricing-refresh-button:hover:not(:disabled) {
  border-color: rgba(96, 165, 250, 0.5);
  color: #93c5fd;
}

.pricing-preset,
.pricing-method-option,
.pricing-stepper,
.pricing-input {
  border-color: rgba(148, 163, 184, 0.42);
  background: rgba(248, 250, 252, 0.9);
  color: #0f172a;
}

.pricing-preset--idle:hover,
.pricing-method-option--idle:hover,
.pricing-stepper-button:hover {
  border-color: rgba(59, 130, 246, 0.5);
  background: #ffffff;
}

.pricing-preset--selected,
.pricing-method-option--selected {
  border-color: rgba(37, 99, 235, 0.74);
  background: rgba(219, 234, 254, 0.76);
  color: #1e3a8a;
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.28);
}

:global(.dark) .pricing-preset,
:global(.dark) .pricing-method-option,
:global(.dark) .pricing-stepper,
:global(.dark) .pricing-input {
  border-color: rgba(51, 65, 85, 0.94);
  background: rgba(15, 23, 42, 0.82);
  color: #f8fafc;
}

:global(.dark) .pricing-preset--idle:hover,
:global(.dark) .pricing-method-option--idle:hover,
:global(.dark) .pricing-stepper-button:hover {
  border-color: rgba(100, 116, 139, 0.92);
  background: rgba(30, 41, 59, 0.86);
}

:global(.dark) .pricing-preset--selected,
:global(.dark) .pricing-method-option--selected {
  border-color: rgba(96, 165, 250, 0.9);
  background: rgba(59, 130, 246, 0.16);
  color: #f8fafc;
  box-shadow: 0 0 0 1px rgba(96, 165, 250, 0.45);
}

.pricing-stepper-value {
  color: #0f172a;
}

:global(.dark) .pricing-stepper-value {
  color: #f8fafc;
}

.pricing-stepper-button {
  color: #64748b;
}

:global(.dark) .pricing-stepper-button {
  color: #94a3b8;
}

.pricing-input::placeholder {
  color: #94a3b8;
}

:global(.dark) .pricing-input::placeholder {
  color: #475569;
}

.pricing-summary {
  background: rgba(239, 246, 255, 0.86);
}

:global(.dark) .pricing-summary {
  background: rgba(18, 28, 43, 0.92);
}

.pricing-divider,
.pricing-section-divider {
  border-color: rgba(148, 163, 184, 0.3);
}

:global(.dark) .pricing-divider,
:global(.dark) .pricing-section-divider {
  border-color: rgba(51, 65, 85, 0.9);
}

.pricing-plan-card--idle:hover {
  border-color: rgba(59, 130, 246, 0.45);
}

.pricing-plan-card--selected {
  border-color: rgba(37, 99, 235, 0.76);
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.28), 0 18px 44px rgba(37, 99, 235, 0.12);
}

.pricing-plan-card--recommended {
  border-color: rgba(99, 102, 241, 0.66);
  box-shadow: 0 0 0 1px rgba(99, 102, 241, 0.2), 0 18px 44px rgba(99, 102, 241, 0.1);
}

:global(.dark) .pricing-plan-card--idle:hover {
  border-color: rgba(100, 116, 139, 0.9);
}

:global(.dark) .pricing-plan-card--selected {
  border-color: rgba(96, 165, 250, 0.9);
  box-shadow: 0 0 0 1px rgba(96, 165, 250, 0.55), 0 18px 48px rgba(0, 0, 0, 0.22);
}

:global(.dark) .pricing-plan-card--recommended {
  border-color: rgba(129, 140, 248, 0.9);
  box-shadow: 0 0 0 1px rgba(129, 140, 248, 0.45), 0 18px 48px rgba(0, 0, 0, 0.22);
}

.pricing-plan-price {
  color: #0f172a;
}

:global(.dark) .pricing-plan-price {
  color: #f8fafc;
}

.pricing-plan-divider,
.pricing-faq-list > div + div {
  border-top: 1px solid rgba(148, 163, 184, 0.28);
}

:global(.dark) .pricing-plan-divider,
:global(.dark) .pricing-faq-list > div + div {
  border-top-color: rgba(51, 65, 85, 0.9);
}

.pricing-feature-list {
  color: #334155;
}

:global(.dark) .pricing-feature-list {
  color: #e2e8f0;
}

.pricing-confirm-panel {
  border-color: rgba(59, 130, 246, 0.36);
  background: rgba(219, 234, 254, 0.56);
}

:global(.dark) .pricing-confirm-panel {
  border-color: rgba(96, 165, 250, 0.42);
  background: rgba(59, 130, 246, 0.12);
}

.pricing-empty-icon {
  color: #94a3b8;
}

:global(.dark) .pricing-empty-icon {
  color: #475569;
}
</style>
