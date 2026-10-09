<template>
  <p v-if="error" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ error }}</p>
  <PaymentQRDialog :show="!!active" :order-id="active?.orderId || 0" :qr-code="active?.qrCode || ''"
    :expires-at="active?.expiresAt || ''" :payment-type="active?.paymentType || ''" :pay-url="active?.payUrl || ''"
    @close="active = null; emit('refresh')" @success="emit('refresh')" />
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref } from 'vue'
import { useRouter } from 'vue-router'
import PaymentQRDialog from '@/components/payment/PaymentQRDialog.vue'
import { decidePaymentLaunch, writePaymentRecoverySnapshot, type PaymentRecoverySnapshot } from '@/components/payment/paymentFlow'
import type { CreateOrderResult } from '@/types/payment'

const router = useRouter()
const emit = defineEmits<{ refresh: [] }>()
const active = ref<PaymentRecoverySnapshot | null>(null)
const error = ref('')
let cancelBridgeWait: (() => void) | undefined
onBeforeUnmount(() => cancelBridgeWait?.())

async function launch(result: CreateOrderResult, requestedMethod = '') {
  error.value = ''
  cancelBridgeWait?.()
  active.value = null
  await nextTick()
  const method = result.payment_type || requestedMethod
  const stripeMethod = method === 'stripe' ? undefined : method === 'wxpay' ? 'wechat_pay' : 'alipay'
  const stripeUrl = result.client_secret ? router.resolve({ path: '/payment/stripe', query: {
    order_id: String(result.order_id), client_secret: result.client_secret,
    resume_token: result.resume_token, method: stripeMethod,
  } }).href : ''
  const airwallexUrl = result.client_secret && result.intent_id ? router.resolve({ path: '/payment/airwallex', query: {
    order_id: String(result.order_id), out_trade_no: result.out_trade_no, resume_token: result.resume_token,
  } }).href : ''
  const decision = decidePaymentLaunch(result, { visibleMethod: method, orderType: 'store',
    isMobile: window.matchMedia('(max-width: 640px)').matches,
    isWechatBrowser: /micromessenger/i.test(navigator.userAgent),
    stripeRouteUrl: stripeUrl, stripePopupUrl: stripeUrl, airwallexRouteUrl: airwallexUrl })
  // OAuth may occur before an order exists. Do not replace recovery with id=0.
  if (result.order_id > 0) writePaymentRecoverySnapshot(localStorage, decision.recovery)
  switch (decision.kind) {
    case 'qr_waiting':
      active.value = decision.paymentState
      return
    case 'redirect_waiting':
    case 'stripe_popup': {
      active.value = decision.paymentState
      const popup = window.open(decision.paymentState.payUrl, '_blank', 'noopener,noreferrer')
      if (!popup) error.value = '付款窗口可能被拦截，请在付款弹窗内点击“打开付款窗口”。'
      return
    }
    case 'stripe_route':
    case 'airwallex_route':
      window.location.assign(decision.paymentState.payUrl)
      return
    case 'wechat_oauth':
      if (decision.oauth?.authorize_url) {
        window.location.assign(decision.oauth.authorize_url)
        return
      }
      break
    case 'wechat_jsapi': {
      const bridge = (window as Window & { WeixinJSBridge?: { invoke: (name: string, data: unknown, callback: (res: { err_msg?: string }) => void) => void } }).WeixinJSBridge
      if (!bridge || !decision.jsapi) {
        error.value = '微信支付环境尚未就绪，请在微信内打开订单继续付款；不要重新创建订单。'
        return
      }
      try {
        const status = await new Promise<string>((resolve) => {
          const finish = (status: string) => { clearTimeout(timer); cancelBridgeWait = undefined; resolve(status) }
          const timer = window.setTimeout(() => finish('timeout'), 60000)
          cancelBridgeWait = () => finish('cancelled')
          bridge.invoke('getBrandWCPayRequest', decision.jsapi, (res) => finish(res.err_msg || ''))
        })
        if (!status.toLowerCase().endsWith(':ok')) error.value = '微信支付未确认完成，请刷新订单状态或继续原订单付款。'
        emit('refresh')
      } catch {
        error.value = '微信支付未确认完成，请前往我的商店订单查看。'
      }
      return
    }
  }
  error.value = '订单已保留，暂未取得付款入口。请到我的商店订单刷新状态或联系管理员，不要重复下单。'
}
defineExpose({ launch })
</script>
