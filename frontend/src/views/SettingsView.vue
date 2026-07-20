<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface Settings {
  totalCapital: number
  positionRatio: number
  stopLossPercent: number
  takeProfitPercent: number
  trailingStopPercent: number
  maxHoldDays: number
  autoScan: boolean
  scanTime: string
}

const SETTINGS_KEY = 'stock_app_settings'

const settings = ref<Settings>({
  totalCapital: 100000,
  positionRatio: 30,
  stopLossPercent: 5,
  takeProfitPercent: 10,
  trailingStopPercent: 3,
  maxHoldDays: 8,
  autoScan: false,
  scanTime: '15:00',
})
const saving = ref(false)

onMounted(() => {
  const saved = localStorage.getItem(SETTINGS_KEY)
  if (saved) {
    try { settings.value = { ...settings.value, ...JSON.parse(saved) } } catch {}
  }
})

const handleSave = () => {
  saving.value = true
  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings.value))
  // Also save stop loss for position page
  localStorage.setItem('stop_loss_percent', String(-settings.value.stopLossPercent))
  setTimeout(() => { saving.value = false }, 500)
}
</script>

<template>
  <div class="bg-page min-h-screen pb-24">
    <div class="max-w-2xl mx-auto px-4 py-6 space-y-5">

      <!-- 资金管理 -->
      <div class="card overflow-hidden">
        <div class="px-5 py-4 border-b border-border flex items-center gap-2.5">
          <svg class="w-5 h-5" style="color:#9AC4DE" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm font-semibold text-text-primary">资金管理</span>
        </div>
        <div class="p-5 space-y-5">
          <div class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">总资金（元）</label>
            <input
              v-model.number="settings.totalCapital"
              type="number"
              placeholder="请输入总资金"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <div class="flex items-center justify-between gap-4">
            <div class="shrink-0">
              <label class="text-sm font-medium text-text-primary">单只仓位比例（%）</label>
            </div>
            <input
              v-model.number="settings.positionRatio"
              type="number"
              placeholder="请输入仓位比例"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <p class="text-xs text-text-tertiary ml-0">建议每只股票的仓位占比，如 30%</p>
        </div>
      </div>

      <!-- 止损止盈策略 -->
      <div class="card overflow-hidden">
        <div class="px-5 py-4 border-b border-border flex items-center gap-2.5">
          <svg class="w-5 h-5" style="color:#7ED6B0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span class="text-sm font-semibold text-text-primary">止损止盈策略</span>
        </div>
        <div class="p-5 space-y-5">
          <div class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">固定止损比例（%）</label>
            <input
              v-model.number="settings.stopLossPercent"
              type="number"
              placeholder="如：5"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <div class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">固定止盈比例（%）</label>
            <input
              v-model.number="settings.takeProfitPercent"
              type="number"
              placeholder="如：10"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <div class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">移动止盈回撤比例（%）</label>
            <input
              v-model.number="settings.trailingStopPercent"
              type="number"
              placeholder="如：3"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <div class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">时间止损（天）</label>
            <input
              v-model.number="settings.maxHoldDays"
              type="number"
              placeholder="如：8"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary num-font outline-none focus:border-brand transition-colors"
            />
          </div>
          <p class="text-xs text-text-tertiary ml-0">持仓超过此天数不涨则考虑卖出</p>
        </div>
      </div>

      <!-- 自动扫描 -->
      <div class="card overflow-hidden">
        <div class="px-5 py-4 border-b border-border flex items-center gap-2.5">
          <svg class="w-5 h-5" style="color:#FFB8B0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <span class="text-sm font-semibold text-text-primary">自动扫描</span>
        </div>
        <div class="p-5 space-y-5">
          <div class="flex items-center justify-between">
            <div>
              <div class="text-sm font-medium text-text-primary">开启自动扫描</div>
              <div class="text-xs text-text-tertiary mt-0.5">每日收盘后自动扫描股票池</div>
            </div>
            <button
              :class="['relative w-11 h-6 rounded-full transition-colors shrink-0', settings.autoScan ? 'bg-brand' : 'bg-border']"
              @click="settings.autoScan = !settings.autoScan"
            >
              <span
                :class="['absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow-sm transition-transform', settings.autoScan ? 'translate-x-[22px]' : 'translate-x-0']"
              />
            </button>
          </div>
          <div v-if="settings.autoScan" class="flex items-center justify-between gap-4">
            <label class="text-sm font-medium text-text-primary shrink-0">扫描时间</label>
            <input
              v-model="settings.scanTime"
              type="text"
              class="w-56 px-3.5 py-2 bg-hover border border-border rounded-lg text-sm text-text-primary placeholder-text-tertiary outline-none focus:border-brand transition-colors"
            />
          </div>
        </div>
      </div>

      <!-- 风险提示 -->
      <div class="card overflow-hidden">
        <div class="p-4 flex items-start gap-3">
          <svg class="w-5 h-5 shrink-0 mt-0.5" style="color:#FFDC9F" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
          </svg>
          <span class="text-xs text-text-secondary leading-relaxed">
            风险提示：本工具仅供参考，不构成投资建议。股市有风险，投资需谨慎。请根据自身风险承受能力合理配置仓位。
          </span>
        </div>
      </div>

    </div>

    <!-- 保存按钮 -->
    <div class="fixed bottom-0 left-0 right-0 bg-white/90 backdrop-blur-sm border-t border-border p-4 z-40">
      <div class="max-w-2xl mx-auto">
        <button
          class="btn btn-brand w-full py-2.5 text-sm font-medium flex items-center justify-center gap-2"
          :disabled="saving"
          @click="handleSave"
        >
          <svg v-if="saving" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          {{ saving ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </div>
  </div>
</template>
