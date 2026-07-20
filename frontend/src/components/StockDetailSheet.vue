<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { getStockDetail, getSellAdvice, getUnifiedScore } from '@/api'
import type { StockTechnicalDetail, SellAdviceResult, StockScore } from '@/types/stock'

const props = defineProps<{
  open: boolean
  code: string
  costPrice?: number
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const detail = ref<StockTechnicalDetail | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const sellAdvice = ref<SellAdviceResult | null>(null)
const unifiedScore = ref<StockScore | null>(null)
const isFavorite = ref(false)

watch(() => [props.open, props.code], () => {
  if (props.open && props.code) {
    if (props.open) fetchDetail()
  }
}, { immediate: true })

const fetchDetail = async () => {
  loading.value = true
  error.value = null
  sellAdvice.value = null
  try {
    const [res, scoreRes] = await Promise.allSettled([
      getStockDetail(props.code),
      getUnifiedScore(props.code),
    ])
    if (res.status === 'fulfilled' && res.value.code === 200 && res.value.data) {
      detail.value = res.value.data as StockTechnicalDetail
    } else {
      error.value = '获取数据失败'
    }
    if (scoreRes.status === 'fulfilled' && scoreRes.value.code === 200 && scoreRes.value.data) {
      unifiedScore.value = scoreRes.value.data as StockScore
    }
    if (props.costPrice && props.costPrice > 0) {
      const sa = await getSellAdvice(props.code, props.costPrice)
      if (sa.code === 200 && sa.data) sellAdvice.value = sa.data
    }
  } catch {
    error.value = '网络请求失败'
  } finally {
    loading.value = false
  }
}

const isPositive = computed(() => (detail.value?.changePercent ?? 0) >= 0)

// MACD chart
const macdChart = computed(() => {
  if (!detail.value?.klineHistory || detail.value.klineHistory.length < 15) return null
  const klines = detail.value.klineHistory.slice(-30)
  const ema = (data: number[], period: number) => {
    const k = 2 / (period + 1)
    const result: number[] = [data[0]]
    for (let i = 1; i < data.length; i++) {
      result.push(data[i] * k + result[i - 1] * (1 - k))
    }
    return result
  }
  const closes = klines.map(k => k.close)
  const ema12 = ema(closes, 12)
  const ema26 = ema(closes, 26)
  const dif = ema12.map((v, i) => v - ema26[i])
  const dea = ema(dif, 9)
  const macd = dif.map((v, i) => 2 * (v - dea[i]))

  const min = Math.min(...dif, ...dea, ...macd)
  const max = Math.max(...dif, ...dea, ...macd)
  const range = max - min || 1

  const w = 100 / (dif.length - 1 || 1)
  const toY = (v: number) => 32 - ((v - min) / range) * 28

  const difPath = dif.map((v, i) => `${i === 0 ? 'M' : 'L'}${(i * w).toFixed(1)},${toY(v).toFixed(1)}`).join(' ')
  const deaPath = dea.map((v, i) => `${i === 0 ? 'M' : 'L'}${(i * w).toFixed(1)},${toY(v).toFixed(1)}`).join(' ')

  return { difPath, deaPath, bars: macd.map((v, i) => ({ value: v, x: i * w, h: Math.max(1, Math.abs(v) / range * 20) })), min, max }
})

// 5-block trend
const trendBlocks = computed(() => {
  if (!detail.value?.klineHistory || detail.value.klineHistory.length < 5) return []
  const closes = detail.value.klineHistory.slice(-5).map(k => k.close)
  return closes.map((c) => {
    const diff = ((c - closes[closes.length - 1]) / closes[closes.length - 1]) * 100
    if (diff > 1) return { up: true, strong: true }
    if (diff > 0) return { up: true, strong: false }
    if (diff > -1) return { up: false, strong: false }
    return { up: false, strong: true }
  })
})

// Boll position
const bollPos = computed(() => {
  if (!detail.value?.boll || !detail.value?.price) return 50
  const { upper, lower } = detail.value.boll
  if (upper === lower) return 50
  return Math.min(100, Math.max(0, ((detail.value.price - lower) / (upper - lower)) * 100))
})

// Sparkline
const sparkline = computed(() => {
  if (!detail.value?.klineHistory || detail.value.klineHistory.length < 3) return null
  const klines = detail.value.klineHistory.slice(-30)
  const prices = klines.map(k => k.close)
  const min = Math.min(...prices)
  const max = Math.max(...prices)
  const range = max - min || 1
  const w = 100 / (klines.length - 1 || 1)
  const points = prices.map((p, i) => `${(i * w).toFixed(1)},${(30 - ((p - min) / range) * 26).toFixed(1)}`)
  return {
    path: `M${points.join(' L')}`,
    up: prices[prices.length - 1] >= prices[0],
  }
})
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="open" class="fixed inset-0 z-50 flex md:justify-end items-end md:items-stretch" @click.self="emit('update:open', false)">
        <div class="absolute inset-0 modal-overlay" @click="emit('update:open', false)" />
        <!-- Mobile: bottom sheet -->
        <div class="md:hidden w-full max-h-[88vh] overflow-y-auto bg-white/95 backdrop-blur-xl rounded-t-3xl shadow-2xl animate-slide-up relative"
          style="border-top: 1px solid rgba(255,255,255,0.3);"
        >
          <!-- Drag handle -->
          <div class="sticky top-0 z-10 bg-white/90 backdrop-blur-md pt-3 pb-1 flex justify-center"
            @click="emit('update:open', false)">
            <div class="w-10 h-1 rounded-full bg-black/20" />
          </div>

          <!-- Header -->
          <div v-if="detail" class="px-5 pt-2 pb-1">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="font-semibold text-text-primary text-lg">{{ detail.name }}</span>
                <span class="text-[10px] text-text-tertiary bg-black/4 px-2 py-0.5 rounded-full">{{ detail.code }}</span>
              </div>
              <button class="w-8 h-8 flex items-center justify-center rounded-2xl hover:bg-black/5 transition-all shrink-0"
                @click="emit('update:open', false)">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="text-text-tertiary"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="flex items-baseline gap-3 mt-1.5 mb-3">
              <span class="text-[28px] font-bold num-font text-text-primary tracking-tight">¥{{ detail.price.toFixed(2) }}</span>
              <span :class="['text-sm font-semibold num-font px-2 py-0.5 rounded-full', isPositive ? 'bg-up-bg text-up' : 'bg-down-bg text-down']">
                <template v-if="isPositive">+</template>{{ detail.changePercent.toFixed(2) }}%
              </span>
            </div>
          </div>

          <!-- Skeleton -->
          <div v-else-if="loading" class="px-5 pb-5 space-y-4">
            <div class="h-5 w-32 skeleton-shimmer rounded-xl" />
            <div class="h-8 w-48 skeleton-shimmer rounded-xl" />
            <div class="h-32 skeleton-shimmer rounded-2xl" />
            <div class="h-32 skeleton-shimmer rounded-2xl" />
          </div>

          <div v-else-if="error" class="px-5 pb-5 text-center">
            <p class="text-sm text-[#A86B62]">{{ error }}</p>
            <button class="text-xs text-earth-sage mt-3 font-medium hover:underline" @click="fetchDetail">重试</button>
          </div>

          <!-- Content -->
          <div v-if="detail" class="px-5 pb-6 space-y-4">
            <!-- Score -->
            <div v-if="unifiedScore" class="flex items-center justify-between bg-white/70 rounded-2xl px-4 py-3.5">
              <span class="text-sm text-text-secondary">综合评分</span>
              <div class="flex items-center gap-3">
                <span class="num-font font-bold text-lg" :class="unifiedScore.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">
                  {{ unifiedScore.totalScore.toFixed(1) }}
                </span>
                <div class="w-14 h-1.5 bg-black/8 rounded-full overflow-hidden">
                  <div class="h-full rounded-full transition-all duration-500" :class="unifiedScore.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(unifiedScore.totalScore, 100) + '%' }" />
                </div>
              </div>
            </div>

            <!-- Sparkline -->
            <div v-if="sparkline" class="card p-4">
              <div class="flex items-center justify-between mb-2">
                <span class="text-xs font-semibold text-text-primary">走势</span>
                <span class="text-[10px] text-text-tertiary">近30日</span>
              </div>
              <svg viewBox="0 0 100 30" class="w-full h-14">
                <path :d="sparkline.path" fill="none" :stroke="sparkline.up ? '#D4837A' : '#7DB89A'" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="sparkline-path" />
              </svg>
            </div>

            <!-- MACD -->
            <div class="card p-4">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">MACD</span>
                <span :class="['text-xs font-medium px-3 py-0.5 rounded-full', detail.macd.dif > detail.macd.dea ? 'bg-up-bg text-up' : 'bg-down-bg text-down']">
                  {{ detail.macd.dif > detail.macd.dea ? '金叉' : '死叉' }}
                </span>
              </div>
              <div class="grid grid-cols-3 gap-2 mb-3">
                <div v-for="(item, i) in [
                  { label: 'DIF', val: detail.macd.dif },
                  { label: 'DEA', val: detail.macd.dea },
                  { label: 'MACD', val: detail.macd.macd },
                ]" :key="i" class="text-center bg-black/4 rounded-xl py-2">
                  <div class="text-[10px] text-text-tertiary mb-0.5">{{ item.label }}</div>
                  <div class="num-font font-semibold text-xs" :class="item.val > 0 ? 'text-up' : 'text-down'">
                    {{ item.val > 0 ? '+' : '' }}{{ item.val.toFixed(3) }}
                  </div>
                </div>
              </div>
              <svg v-if="macdChart" viewBox="0 0 100 32" class="w-full h-16">
                <line x1="0" y1="16" x2="100" y2="16" stroke="rgba(0,0,0,0.06)" stroke-width="0.5" />
                <rect v-for="bar in macdChart.bars" :key="bar.x" :x="bar.x" :y="16 - (bar.value >= 0 ? bar.h : 0)" width="2" :height="bar.h" :fill="bar.value >= 0 ? '#F5E3E0' : '#E3F0EA'" :opacity="0.8" />
                <path :d="macdChart.difPath" fill="none" stroke="#8DB8A0" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                <path :d="macdChart.deaPath" fill="none" stroke="#9A9A9A" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" stroke-dasharray="3 2" />
              </svg>
            </div>

            <!-- BOLL -->
            <div class="card p-4">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">BOLL</span>
                <span class="text-[10px] text-text-tertiary bg-black/4 px-3 py-0.5 rounded-full">{{ detail.boll.position }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2 mb-3">
                <div v-for="(item, i) in [
                  { label: '上轨', val: detail.boll.upper, cls: 'text-text-secondary' },
                  { label: '中轨', val: detail.boll.middle, cls: 'text-text-primary' },
                  { label: '下轨', val: detail.boll.lower, cls: 'text-text-secondary' },
                ]" :key="i" class="text-center bg-black/4 rounded-xl py-2">
                  <div class="text-[10px] text-text-tertiary mb-0.5">{{ item.label }}</div>
                  <div class="num-font font-medium text-xs text-text-primary">{{ item.val.toFixed(2) }}</div>
                </div>
              </div>
              <div class="relative h-2 bg-gradient-to-r from-earth-sage-light via-earth-sky-light to-earth-terracotta-light rounded-full">
                <div
                  class="absolute top-1/2 -translate-y-1/2 w-3.5 h-3.5 bg-white rounded-full shadow-md transition-all duration-300"
                  style="border: 2px solid #8DB8A0;"
                  :style="{ left: `calc(${bollPos}% - 7px)` }"
                />
              </div>
              <div class="text-[10px] text-text-tertiary text-center mt-2">带宽 {{ detail.boll.bandwidth.toFixed(2) }}%</div>
            </div>

            <!-- Trend prediction -->
            <div v-if="trendBlocks.length > 0" class="card p-4">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">走势预测</span>
                <span class="text-[10px] text-text-tertiary">近5日</span>
              </div>
              <div class="flex gap-2 justify-center">
                <div
                  v-for="(block, i) in trendBlocks"
                  :key="i"
                  class="w-9 h-9 rounded-xl flex items-center justify-center text-xs font-bold transition-all duration-300 spring-up"
                  :style="{ animationDelay: `${i * 0.08}s` }"
                  :class="block.up
                    ? (block.strong ? 'bg-up-bg text-up' : 'bg-up-bg/50 text-up/60')
                    : (block.strong ? 'bg-down-bg text-down' : 'bg-down-bg/50 text-down/60')"
                >
                  <svg v-if="block.up" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="18 15 12 9 6 15"/></svg>
                  <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
                </div>
              </div>
            </div>

            <!-- Sell advice -->
            <div v-if="sellAdvice" class="card p-4">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">止盈参考</span>
                <span class="num-font text-sm font-semibold" :class="(sellAdvice as any).profitPercent >= 0 ? 'text-up' : 'text-down'">
                  {{ (sellAdvice as any).profitPercent >= 0 ? '+' : '' }}{{ (sellAdvice as any).profitPercent.toFixed(2) }}%
                </span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <div v-for="(t, i) in [(sellAdvice as any).target1, (sellAdvice as any).target2]" :key="i" class="bg-white/70 rounded-xl p-2.5 text-center">
                  <div class="text-[10px] text-text-tertiary mb-1">{{ t.label }}</div>
                  <div class="num-font font-semibold text-xs text-text-primary">¥{{ t.price.toFixed(2) }}</div>
                  <div class="text-xs num-font font-medium mt-0.5" :class="t.profit >= 0 ? 'text-up' : 'text-down'">{{ t.profit >= 0 ? '+' : '' }}{{ t.profit.toFixed(1) }}%</div>
                </div>
                <div class="bg-up-bg/30 rounded-xl p-2.5 text-center">
                  <div class="text-[10px] text-text-tertiary mb-1">止损</div>
                  <div class="num-font font-semibold text-xs text-[#A86B62]">¥{{ (sellAdvice as any).stopLossPrice.toFixed(2) }}</div>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex gap-3 pt-1 pb-2">
              <button class="flex-1 py-2.5 rounded-2xl bg-black/5 text-xs text-text-secondary hover:bg-black/10 hover:text-text-primary transition-all duration-300"
                @click="isFavorite = !isFavorite">
                <svg width="13" height="13" viewBox="0 0 24 24" :fill="isFavorite ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" class="inline mr-1 align-middle"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                关注
              </button>
              <button class="flex-1 py-2.5 rounded-2xl text-xs text-text-tertiary/50 transition-all duration-300 cursor-not-allowed" disabled>
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="inline mr-1 align-middle"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                笔记
              </button>
              <button class="flex-1 py-2.5 rounded-2xl text-xs text-text-tertiary/50 transition-all duration-300 cursor-not-allowed" disabled>
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="inline mr-1 align-middle"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
                提醒
              </button>
            </div>
          </div>
        </div>

        <!-- Desktop: right side drawer -->
        <div class="hidden md:block relative w-[420px] max-w-[92vw] h-full overflow-y-auto bg-white/90 backdrop-blur-xl shadow-2xl"
          style="border-left: 1px solid rgba(255,255,255,0.3);">
          <!-- Header -->
          <div v-if="detail" class="sticky top-0 bg-white/80 backdrop-blur-md px-6 py-5 z-10"
            style="border-bottom: 1px solid rgba(0,0,0,0.04);">
            <div class="flex items-center justify-between">
              <div>
                <div class="flex items-center gap-2.5">
                  <span class="font-semibold text-text-primary text-lg">{{ detail.name }}</span>
                  <span class="text-xs text-text-tertiary bg-black/4 px-2 py-0.5 rounded-full">{{ detail.code }}</span>
                </div>
                <div class="flex items-baseline gap-3 mt-1.5">
                  <span class="text-[28px] font-bold num-font text-text-primary tracking-tight">¥{{ detail.price.toFixed(2) }}</span>
                  <span :class="['text-sm font-semibold num-font px-2 py-0.5 rounded-full', isPositive ? 'bg-up-bg text-up' : 'bg-down-bg text-down']">
                    <template v-if="isPositive">+</template>{{ detail.changePercent.toFixed(2) }}%
                  </span>
                </div>
              </div>
              <button
                class="w-9 h-9 flex items-center justify-center rounded-2xl hover:bg-black/5 transition-all duration-300 shrink-0"
                @click="emit('update:open', false)"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="text-text-tertiary"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
          </div>

          <!-- Desktop skeleton -->
          <div v-else-if="loading" class="p-6 space-y-5">
            <div class="h-5 w-32 skeleton-shimmer rounded-xl" />
            <div class="h-8 w-48 skeleton-shimmer rounded-xl" />
            <div class="h-36 skeleton-shimmer rounded-2xl" />
            <div class="h-36 skeleton-shimmer rounded-2xl" />
          </div>

          <div v-else-if="error" class="p-6 text-center">
            <p class="text-sm text-[#A86B62]">{{ error }}</p>
            <button class="text-xs text-earth-sage mt-3 font-medium hover:underline" @click="fetchDetail">重试</button>
          </div>

          <!-- Desktop content -->
          <div v-if="detail" class="p-6 space-y-5">
            <!-- Score -->
            <div v-if="unifiedScore" class="flex items-center justify-between bg-white/70 rounded-2xl px-5 py-4">
              <span class="text-sm text-text-secondary">综合评分</span>
              <div class="flex items-center gap-3">
                <span class="num-font font-bold text-lg" :class="unifiedScore.totalScore >= 60 ? 'text-earth-sage' : 'text-text-tertiary'">
                  {{ unifiedScore.totalScore.toFixed(1) }}
                </span>
                <div class="w-16 h-1.5 bg-black/8 rounded-full overflow-hidden">
                  <div class="h-full rounded-full transition-all duration-500" :class="unifiedScore.totalScore >= 60 ? 'bg-earth-sage' : 'bg-text-tertiary'" :style="{ width: Math.min(unifiedScore.totalScore, 100) + '%' }" />
                </div>
              </div>
            </div>

            <!-- Sparkline -->
            <div v-if="sparkline" class="card p-5">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">走势</span>
                <span class="text-xs text-text-tertiary">近30日</span>
              </div>
              <svg viewBox="0 0 100 30" class="w-full h-16">
                <path :d="sparkline.path" fill="none" :stroke="sparkline.up ? '#D4837A' : '#7DB89A'" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="sparkline-path" />
              </svg>
            </div>

            <!-- MACD -->
            <div class="card p-5">
              <div class="flex items-center justify-between mb-4">
                <span class="text-xs font-semibold text-text-primary">MACD</span>
                <span :class="['text-xs font-medium px-3 py-1 rounded-full', detail.macd.dif > detail.macd.dea ? 'bg-up-bg text-up' : 'bg-down-bg text-down']">
                  {{ detail.macd.dif > detail.macd.dea ? '金叉' : '死叉' }}
                </span>
              </div>
              <div class="grid grid-cols-3 gap-3 mb-4">
                <div v-for="(item, i) in [
                  { label: 'DIF', val: detail.macd.dif },
                  { label: 'DEA', val: detail.macd.dea },
                  { label: 'MACD', val: detail.macd.macd },
                ]" :key="i" class="text-center bg-black/4 rounded-xl py-2.5">
                  <div class="text-xs text-text-tertiary mb-1">{{ item.label }}</div>
                  <div class="num-font font-semibold text-sm" :class="item.val > 0 ? 'text-up' : 'text-down'">
                    {{ item.val > 0 ? '+' : '' }}{{ item.val.toFixed(3) }}
                  </div>
                </div>
              </div>
              <svg v-if="macdChart" viewBox="0 0 100 32" class="w-full h-20">
                <line x1="0" y1="16" x2="100" y2="16" stroke="rgba(0,0,0,0.06)" stroke-width="0.5" />
                <rect v-for="bar in macdChart.bars" :key="bar.x" :x="bar.x" :y="16 - (bar.value >= 0 ? bar.h : 0)" width="2" :height="bar.h" :fill="bar.value >= 0 ? '#F5E3E0' : '#E3F0EA'" :opacity="0.8" />
                <path :d="macdChart.difPath" fill="none" stroke="#8DB8A0" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
                <path :d="macdChart.deaPath" fill="none" stroke="#9A9A9A" stroke-width="1" stroke-linecap="round" stroke-linejoin="round" stroke-dasharray="3 2" />
              </svg>
              <div class="flex items-center gap-4 mt-3 text-[10px] text-text-tertiary">
                <span class="flex items-center gap-1.5"><span class="w-3 h-0.5 bg-earth-sage inline-block rounded-full" /> DIF</span>
                <span class="flex items-center gap-1.5"><span class="w-3 h-0 inline-block" style="border-top: 1.5px dashed #9A9A9A;" /> DEA</span>
              </div>
            </div>

            <!-- BOLL -->
            <div class="card p-5">
              <div class="flex items-center justify-between mb-4">
                <span class="text-xs font-semibold text-text-primary">BOLL</span>
                <span class="text-xs text-text-tertiary bg-black/4 px-3 py-1 rounded-full">{{ detail.boll.position }}</span>
              </div>
              <div class="grid grid-cols-3 gap-3 mb-4">
                <div v-for="(item, i) in [
                  { label: '上轨', val: detail.boll.upper, cls: 'text-text-secondary' },
                  { label: '中轨', val: detail.boll.middle, cls: 'text-text-primary' },
                  { label: '下轨', val: detail.boll.lower, cls: 'text-text-secondary' },
                ]" :key="i" class="text-center bg-black/4 rounded-xl py-2.5">
                  <div class="text-xs text-text-tertiary mb-1">{{ item.label }}</div>
                  <div class="num-font font-medium text-sm text-text-primary">{{ item.val.toFixed(2) }}</div>
                </div>
              </div>
              <div class="relative h-2 bg-gradient-to-r from-earth-sage-light via-earth-sky-light to-earth-terracotta-light rounded-full">
                <div
                  class="absolute top-1/2 -translate-y-1/2 w-3.5 h-3.5 bg-white rounded-full shadow-md transition-all duration-300"
                  style="border: 2px solid #8DB8A0;"
                  :style="{ left: `calc(${bollPos}% - 7px)` }"
                />
              </div>
              <div class="text-xs text-text-tertiary text-center mt-2">带宽 {{ detail.boll.bandwidth.toFixed(2) }}%</div>
            </div>

            <!-- Trend prediction -->
            <div v-if="trendBlocks.length > 0" class="card p-5">
              <div class="flex items-center justify-between mb-4">
                <span class="text-xs font-semibold text-text-primary">走势预测</span>
                <span class="text-xs text-text-tertiary">近5日</span>
              </div>
              <div class="flex gap-2.5 justify-center">
                <div
                  v-for="(block, i) in trendBlocks"
                  :key="i"
                  class="w-10 h-10 rounded-xl flex items-center justify-center text-xs font-bold transition-all duration-300 spring-up"
                  :style="{ animationDelay: `${i * 0.08}s` }"
                  :class="block.up
                    ? (block.strong ? 'bg-up-bg text-up' : 'bg-up-bg/50 text-up/60')
                    : (block.strong ? 'bg-down-bg text-down' : 'bg-down-bg/50 text-down/60')"
                >
                  <svg v-if="block.up" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="18 15 12 9 6 15"/></svg>
                  <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><polyline points="6 9 12 15 18 9"/></svg>
                </div>
              </div>
            </div>

            <!-- Sell advice -->
            <div v-if="sellAdvice" class="card p-5">
              <div class="flex items-center justify-between mb-3">
                <span class="text-xs font-semibold text-text-primary">止盈参考</span>
                <span class="num-font text-sm font-semibold" :class="(sellAdvice as any).profitPercent >= 0 ? 'text-up' : 'text-down'">
                  {{ (sellAdvice as any).profitPercent >= 0 ? '+' : '' }}{{ (sellAdvice as any).profitPercent.toFixed(2) }}%
                </span>
              </div>
              <div class="flex gap-3">
                <div v-for="(t, i) in [(sellAdvice as any).target1, (sellAdvice as any).target2]" :key="i" class="flex-1 bg-white/70 rounded-xl p-3 text-center">
                  <div class="text-xs text-text-tertiary mb-1.5">{{ t.label }}</div>
                  <div class="num-font font-semibold text-sm text-text-primary">¥{{ t.price.toFixed(2) }}</div>
                  <div class="text-xs num-font font-medium mt-0.5" :class="t.profit >= 0 ? 'text-up' : 'text-down'">{{ t.profit >= 0 ? '+' : '' }}{{ t.profit.toFixed(1) }}%</div>
                </div>
                <div class="flex-1 bg-up-bg/30 rounded-xl p-3 text-center">
                  <div class="text-xs text-text-tertiary mb-1.5">止损</div>
                  <div class="num-font font-semibold text-sm text-[#A86B62]">¥{{ (sellAdvice as any).stopLossPrice.toFixed(2) }}</div>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex gap-3 pt-2">
              <button class="flex-1 py-2.5 rounded-2xl bg-black/5 text-sm text-text-secondary hover:bg-black/10 hover:text-text-primary transition-all duration-300"
                @click="isFavorite = !isFavorite">
                <svg width="14" height="14" viewBox="0 0 24 24" :fill="isFavorite ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" class="inline mr-1.5 align-middle"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
                关注
              </button>
              <button class="flex-1 py-2.5 rounded-2xl text-sm text-text-tertiary/50 transition-all duration-300 cursor-not-allowed" disabled>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="inline mr-1.5 align-middle"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                笔记
              </button>
              <button class="flex-1 py-2.5 rounded-2xl text-sm text-text-tertiary/50 transition-all duration-300 cursor-not-allowed" disabled>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" class="inline mr-1.5 align-middle"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>
                提醒
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>
