<script setup lang="ts">
import { ref } from 'vue'
import { diagnoseStock, scanCoreSatellite, getWatchList, addWatchList, removeWatchList, scanWatchList } from '@/api'
import type { ScanResult, WatchScanResult, DiagnoseResult, WatchScanStock } from '@/types/stock'
import StockDetailSheet from '@/components/StockDetailSheet.vue'
import StockTag from '@/components/StockTag.vue'
import { getRecommendationStyle } from '@/utils/tagColors'

const inputCode = ref('')
const watchCode = ref('')
const diagnosing = ref(false)
const diagnoseResult = ref<DiagnoseResult | null>(null)
const scanLoading = ref(false)
const watchScanLoading = ref(false)
const scanResult = ref<ScanResult | null>(null)
const watchScanResult = ref<WatchScanResult | null>(null)
const scanError = ref<'empty' | 'timeout' | 'error' | null>(null)
const watchList = ref<string[]>([])
const detailOpen = ref(false)
const selectedCode = ref('')
const alertVisible = ref(false)
const alertStock = ref<WatchScanStock | null>(null)

const handleDiagnose = async () => {
  if (!inputCode.value.trim()) return
  diagnosing.value = true
  try {
    const res = await diagnoseStock(inputCode.value.trim())
    if (res.code === 200) diagnoseResult.value = res.data as DiagnoseResult
  } catch (e) { console.error('诊断失败:', e) } finally { diagnosing.value = false }
}

const handleFullScan = async () => {
  scanLoading.value = true; scanResult.value = null; scanError.value = null
  try {
    const res = await scanCoreSatellite()
    if (res.code === 200) {
      const data = res.data as ScanResult
      scanResult.value = data
      if (!data || (data.core.length === 0 && data.satellite.length === 0)) scanError.value = 'empty'
    } else { scanError.value = 'error' }
  } catch (error: unknown) {
    if (error instanceof Error) scanError.value = error.name === 'AbortError' ? 'timeout' : 'error'
    else scanError.value = 'error'
  } finally { scanLoading.value = false }
}

const handleAddWatch = async () => {
  if (!watchCode.value.trim()) return
  try {
    const res = await addWatchList(watchCode.value.trim())
    if (res.code === 200) { watchList.value.push(watchCode.value.trim()); watchCode.value = '' }
  } catch (e) { console.error('添加关注失败:', e) }
}

const handleRemoveWatch = async (code: string) => {
  try { await removeWatchList(code); watchList.value = watchList.value.filter(c => c !== code) }
  catch (e) { console.error('移除失败:', e) }
}

const handleScanWatchList = async () => {
  watchScanLoading.value = true
  try {
    const res = await scanWatchList()
    if (res.code === 200) {
      const result = res.data as WatchScanResult
      watchScanResult.value = result
      if (result.signals?.length > 0) { alertStock.value = result.signals[0]; alertVisible.value = true }
    }
  } catch (e) { console.error('扫描失败:', e) } finally { watchScanLoading.value = false }
}

const loadWatchList = async () => {
  try { const res = await getWatchList(); if (res.code === 200) watchList.value = res.data as string[] }
  catch (e) { console.error('获取关注列表失败:', e) }
}

const openDetail = (code: string) => { selectedCode.value = code; detailOpen.value = true }

const formatScore = (score: number) => {
  const intScore = Math.floor(score); const decimal = score - intScore
  return decimal > 0.01 ? score.toFixed(1) : intScore.toString()
}

const getDiagnosisRecText = (rec: string): string => {
  switch (rec) { case 'buy': return '建议买入'; case 'avoid': return '建议规避'; default: return '建议观望' }
}

const scoreBarColor = (score: number): string => {
  if (score >= 70) return 'bg-up'
  if (score >= 50) return 'bg-earth-clay-light'
  return 'bg-earth-sage-light'
}

loadWatchList()
</script>

<template>
  <div class="space-y-5">
    <!-- ═══ 个股诊断 — 独立整行 ═══ -->
    <div class="card p-6 space-y-4">
      <div class="flex items-center gap-2.5">
        <span class="text-sm font-semibold text-text-primary">个股诊断</span>
        <span class="text-[10px] text-text-tertiary bg-black/4 px-2 py-0.5 rounded-full">全指标分析</span>
      </div>
      <div class="flex gap-2.5">
        <div class="flex-1 flex items-center bg-black/4 rounded-2xl px-4 py-2.5">
          <svg class="w-4 h-4 text-text-tertiary shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
          <input v-model="inputCode" class="flex-1 ml-2.5 bg-transparent outline-none text-sm text-text-primary placeholder:text-text-tertiary" placeholder="输入股票代码" @keyup.enter="handleDiagnose" />
        </div>
        <button class="btn btn-primary px-5 py-2.5 text-sm font-medium" :class="{ 'opacity-50': diagnosing }" :disabled="diagnosing" @click="handleDiagnose">{{ diagnosing ? '诊断中...' : '诊断' }}</button>
      </div>
      <Transition name="page">
        <div v-if="diagnoseResult" class="space-y-3">
          <!-- 头部：名称+评分+推荐 -->
          <div class="flex items-center justify-between bg-white/70 rounded-2xl px-5 py-3.5">
            <div class="flex items-center gap-3">
              <span class="font-bold text-text-primary text-base">{{ diagnoseResult.name }}</span>
              <span class="text-xs text-text-tertiary">{{ diagnoseResult.code }}</span>
              <span class="num-font text-sm font-semibold" :class="diagnoseResult.changePercent >= 0 ? 'text-up' : 'text-down'">
                {{ diagnoseResult.changePercent >= 0 ? '+' : '' }}{{ diagnoseResult.changePercent.toFixed(2) }}%
              </span>
              <span class="num-font text-xs text-text-tertiary">¥{{ diagnoseResult.price.toFixed(2) }}</span>
            </div>
            <div class="flex items-center gap-3">
              <span class="num-font text-lg font-bold" :class="diagnoseResult.score >= 60 ? 'text-down' : diagnoseResult.score >= 30 ? 'text-earth-clay' : 'text-up'">{{ diagnoseResult.score.toFixed(1) }}</span>
              <span :class="['text-xs font-semibold px-3 py-1 rounded-full', getRecommendationStyle(diagnoseResult.recLabel)]">{{ diagnoseResult.recLabel }}</span>
            </div>
          </div>

          <!-- 亮点标签 -->
          <div v-if="diagnoseResult.highlights.length" class="flex flex-wrap gap-1.5">
            <span v-for="h in diagnoseResult.highlights" :key="h"
              class="text-[10px] px-2.5 py-1 rounded-full font-medium"
              :style="{ background: '#E3EDF5', color: '#4A7A9E' }"
            >{{ h }}</span>
          </div>

          <!-- 指标网格 -->
          <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
            <!-- 四维评分 -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">评分</div>
              <div class="space-y-1.5">
                <div class="flex justify-between text-xs"><span>趋势</span><span class="num-font font-semibold">{{ diagnoseResult.trendScore.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>动量</span><span class="num-font font-semibold">{{ diagnoseResult.momentumScore.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>量能</span><span class="num-font font-semibold">{{ diagnoseResult.volumeScore.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>技术</span><span class="num-font font-semibold">{{ diagnoseResult.techScore.toFixed(1) }}</span></div>
              </div>
            </div>

            <!-- MACD -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">MACD</div>
              <div class="space-y-1">
                <div class="flex justify-between text-xs"><span>DIF</span><span class="num-font">{{ diagnoseResult.macdDif.toFixed(3) }}</span></div>
                <div class="flex justify-between text-xs"><span>DEA</span><span class="num-font">{{ diagnoseResult.macdDea.toFixed(3) }}</span></div>
                <div class="flex justify-between text-xs"><span>柱值</span><span class="num-font">{{ diagnoseResult.macdHist.toFixed(3) }}</span></div>
                <div class="flex justify-between text-xs"><span>信号</span><span class="font-semibold text-xs">{{ diagnoseResult.macdSignal || '--' }}</span></div>
              </div>
            </div>

            <!-- BOLL -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">布林带</div>
              <div class="space-y-1">
                <div class="flex justify-between text-xs"><span>上轨</span><span class="num-font">{{ diagnoseResult.bollUpper.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>中轨</span><span class="num-font">{{ diagnoseResult.bollMiddle.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>下轨</span><span class="num-font">{{ diagnoseResult.bollLower.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>位置</span><span class="font-semibold text-xs">{{ diagnoseResult.bollPos || '--' }}</span></div>
              </div>
            </div>

            <!-- 均线 -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">均线</div>
              <div class="space-y-1">
                <div class="flex justify-between text-xs"><span>MA5</span><span class="num-font">{{ diagnoseResult.ma5.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>MA10</span><span class="num-font">{{ diagnoseResult.ma10.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>MA20</span><span class="num-font">{{ diagnoseResult.ma20.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>MA60</span><span class="num-font">{{ diagnoseResult.ma60.toFixed(2) }}</span></div>
              </div>
            </div>

            <!-- RSI / KDJ -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">RSI / KDJ</div>
              <div class="space-y-1">
                <div class="flex justify-between text-xs"><span>RSI(14)</span><span class="num-font font-semibold" :class="diagnoseResult.rsi > 70 ? 'text-up' : diagnoseResult.rsi < 30 ? 'text-down' : ''">{{ diagnoseResult.rsi.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>KDJ-K</span><span class="num-font">{{ diagnoseResult.kdjK.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>KDJ-D</span><span class="num-font">{{ diagnoseResult.kdjD.toFixed(1) }}</span></div>
                <div class="flex justify-between text-xs"><span>KDJ-J</span><span class="num-font">{{ diagnoseResult.kdjJ.toFixed(1) }}</span></div>
              </div>
            </div>

            <!-- 量能 + 箱体 -->
            <div class="bg-white/60 rounded-2xl p-3.5 space-y-2">
              <div class="text-[10px] font-medium text-text-tertiary uppercase tracking-wide">量能 / 支撑</div>
              <div class="space-y-1">
                <div class="flex justify-between text-xs"><span>量比</span><span class="num-font">{{ diagnoseResult.volumeRatio.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>换手</span><span class="num-font">{{ diagnoseResult.turnoverRate.toFixed(2) }}%</span></div>
                <div class="flex justify-between text-xs"><span>支撑</span><span class="num-font">¥{{ diagnoseResult.support.toFixed(2) }}</span></div>
                <div class="flex justify-between text-xs"><span>阻力</span><span class="num-font">¥{{ diagnoseResult.resistance.toFixed(2) }}</span></div>
              </div>
            </div>
          </div>

          <!-- 箱体（有条件） -->
          <div v-if="diagnoseResult.boxBottom > 0" class="bg-white/60 rounded-2xl p-3.5">
            <div class="text-[10px] font-medium text-text-tertiary mb-2 uppercase tracking-wide">箱体形态</div>
            <div class="flex items-center gap-4">
              <div class="flex-1">
                <div class="relative h-4 bg-black/8 rounded-full overflow-hidden">
                  <div class="absolute inset-y-0 left-0 rounded-full transition-all"
                    :style="{ width: diagnoseResult.boxPosPct + '%', background: diagnoseResult.boxPosPct <= 20 ? '#7DB89A' : diagnoseResult.boxPosPct >= 80 ? '#D4837A' : '#8DB8A0' }">
                  </div>
                </div>
                <div class="flex justify-between text-[10px] text-text-tertiary mt-1">
                  <span>箱底 ¥{{ diagnoseResult.boxBottom.toFixed(2) }}</span>
                  <span class="num-font font-semibold" style="color:#2C2C2C">{{ diagnoseResult.boxPosPct.toFixed(0) }}%</span>
                  <span>箱顶 ¥{{ diagnoseResult.boxTop.toFixed(2) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 分析结论 -->
          <div class="text-sm text-text-secondary leading-relaxed bg-black/4 rounded-2xl px-4 py-3">{{ diagnoseResult.analysis }}</div>
        </div>
      </Transition>
    </div>

    <!-- ═══ 关注列表 + 全池扫描 — 左右并列 ═══ -->
    <div class="card-grid md:card-grid-2">
      <!-- 关注列表 -->
      <div class="card p-5 space-y-4">
        <div class="text-sm font-semibold text-text-primary">关注列表</div>
        <div class="flex gap-2.5">
          <div class="flex-1 flex items-center bg-black/4 rounded-2xl px-4 py-2.5">
            <svg class="w-4 h-4 text-text-tertiary shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 4v16m8-8H4" /></svg>
            <input v-model="watchCode" class="flex-1 ml-2.5 bg-transparent outline-none text-sm text-text-primary placeholder:text-text-tertiary" placeholder="输入代码" />
          </div>
          <button class="btn btn-primary px-4 py-2.5 text-sm font-medium" @click="handleAddWatch">添加</button>
        </div>
        <div v-if="watchList.length > 0" class="space-y-1.5">
          <div v-for="code in watchList" :key="code" class="flex items-center justify-between p-2.5 bg-white/70 rounded-2xl">
            <span class="text-sm text-text-primary font-medium">{{ code }}</span>
            <div class="flex gap-1.5 items-center">
              <button class="btn-ghost text-xs px-2.5 py-1 rounded-xl" @click="inputCode = code; handleDiagnose()">诊断</button>
              <button class="btn-ghost text-xs px-2 py-1 rounded-xl text-text-tertiary hover:text-up transition-all" @click="handleRemoveWatch(code)">
                <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
              </button>
            </div>
          </div>
          <button class="w-full btn btn-primary py-2.5 text-sm font-medium mt-2" :class="{ 'opacity-50': watchScanLoading }" :disabled="watchScanLoading" @click="handleScanWatchList">
            {{ watchScanLoading ? '扫描中...' : '扫描关注' }}
          </button>
        </div>
        <div v-else class="text-sm text-text-tertiary text-center py-4">暂无关注的股票</div>
      </div>

      <!-- 全池扫描 -->
      <div class="card p-5 flex flex-col items-center justify-center gap-4 text-center"
        style="border: 1.5px dashed rgba(0,0,0,0.06); background: rgba(255,255,255,0.45); min-height: 200px;">
        <svg class="w-8 h-8 text-text-tertiary/50" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
        <div>
          <div class="text-sm font-medium text-text-secondary">全池扫描</div>
          <div class="text-[11px] text-text-tertiary mt-1">核心 + 卫星双引擎筛选</div>
        </div>
        <button class="btn btn-primary px-6 py-2.5 text-sm font-medium" :class="{ 'opacity-50': scanLoading }" :disabled="scanLoading" @click="handleFullScan">
          {{ scanLoading ? '扫描中...' : '开始扫描' }}
        </button>
      </div>
    </div>

    <!-- ═══ 错误/空状态 — 小尺寸 ═══ -->
    <div v-if="!scanLoading && scanError === 'timeout'" class="flex items-center gap-2.5 bg-[#F5EDDF]/60 rounded-2xl px-4 py-3 text-sm text-[#A8884A]">
      <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>扫描超时，服务繁忙，稍后重试</span>
      <button class="ml-auto text-xs font-medium underline" @click="scanError = null">关闭</button>
    </div>
    <div v-else-if="!scanLoading && scanError === 'error'" class="flex items-center gap-2.5 bg-[#F5E3E0]/60 rounded-2xl px-4 py-3 text-sm text-[#A86B62]">
      <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>
      <span>扫描服务异常，检查控制台日志</span>
      <button class="ml-auto text-xs font-medium underline" @click="scanError = null">关闭</button>
    </div>
    <div v-else-if="!scanLoading && scanError === 'empty'" class="flex items-center gap-2.5 bg-black/4 rounded-2xl px-4 py-3 text-sm text-text-tertiary">
      <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
      <span>当前策略未匹配到符合条件的股票</span>
      <span class="ml-auto text-xs">放宽条件或稍后再试</span>
    </div>

    <!-- ═══ 扫描结果 — 单列满宽 ═══ -->
    <template v-if="scanResult && scanError !== 'empty'">
      <!-- 核心池 -->
      <div v-if="scanResult.core.length > 0" class="space-y-3">
        <div class="flex items-center gap-2.5">
          <div class="w-1 h-5 rounded-full bg-earth-sage" />
          <span class="text-sm font-semibold text-text-primary">核心标的</span>
          <span class="tag-organic text-[10px] bg-earth-sage-light text-earth-sage">仓位 {{ (scanResult.coreTotalWeight * 100).toFixed(0) }}%</span>
        </div>
        <div v-for="(stock, index) in scanResult.core" :key="stock.code"
          class="card p-5 cursor-pointer spring-up"
          :style="{ animationDelay: `${index * 0.04}s` }"
          @click="openDetail(stock.code)">
          <div class="flex justify-between items-start mb-3">
            <div>
              <span class="font-bold text-text-primary text-base">{{ stock.name }}</span>
              <span class="text-xs text-text-tertiary ml-2">{{ stock.code }}</span>
            </div>
            <div class="text-right">
              <div class="font-bold text-text-primary num-font text-lg">¥{{ stock.price.toFixed(2) }}</div>
              <div :class="['text-sm font-medium num-font', stock.changePercent >= 0 ? 'text-up' : 'text-down']">{{ stock.changePercent >= 0 ? '+' : '' }}{{ stock.changePercent.toFixed(2) }}%</div>
            </div>
          </div>
          <div class="flex items-center gap-2 flex-wrap mb-3">
            <StockTag type="highlight" :text="stock.sector" />
            <div class="flex items-center gap-1.5">
              <span class="text-xs font-semibold text-text-primary num-font">{{ formatScore(stock.score.totalScore) }}</span>
              <div class="w-16 h-1 bg-black/8 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all" :class="scoreBarColor(stock.score.totalScore)" :style="{ width: Math.min(stock.score.totalScore, 100) + '%' }" />
              </div>
            </div>
            <StockTag v-if="stock.recommendation" type="recommendation" :text="stock.recommendation" />
          </div>
          <div class="flex items-center justify-between bg-white/70 rounded-xl px-4 py-3">
            <span class="text-xs text-text-secondary">建议仓位</span>
            <span class="text-lg font-bold num-font text-earth-sage">{{ (stock.recommendedPosition * 100).toFixed(1) }}%</span>
          </div>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <StockTag v-if="stock.macdSignal" type="macdSignal" :text="stock.macdSignal" />
            <StockTag v-if="stock.bollPosition" type="bollPosition" :text="stock.bollPosition" />
          </div>
        </div>
      </div>

      <!-- 卫星池 -->
      <div v-if="scanResult.satellite.length > 0" class="space-y-3">
        <div class="flex items-center gap-2.5">
          <div class="w-1 h-5 rounded-full bg-earth-sky" />
          <span class="text-sm font-semibold text-text-primary">卫星标的</span>
          <span class="tag-organic text-[10px] bg-earth-sky-light text-earth-sky">仓位 {{ (scanResult.satelliteTotalWeight * 100).toFixed(0) }}%</span>
        </div>
        <div v-for="stock in scanResult.satellite" :key="stock.code"
          class="card p-5 cursor-pointer" @click="openDetail(stock.code)">
          <div class="flex justify-between items-start mb-3">
            <div>
              <span class="font-bold text-text-primary text-base">{{ stock.name }}</span>
              <span class="text-xs text-text-tertiary ml-2">{{ stock.code }}</span>
            </div>
            <div class="text-right">
              <div class="font-bold text-text-primary num-font text-lg">¥{{ stock.price.toFixed(2) }}</div>
              <div :class="['text-sm font-medium num-font', stock.changePercent >= 0 ? 'text-up' : 'text-down']">{{ stock.changePercent >= 0 ? '+' : '' }}{{ stock.changePercent.toFixed(2) }}%</div>
            </div>
          </div>
          <div class="flex items-center gap-2 flex-wrap mb-3">
            <StockTag type="highlight" :text="stock.sector" />
            <div class="flex items-center gap-1.5">
              <span class="text-xs font-semibold text-text-primary num-font">{{ formatScore(stock.score.totalScore) }}</span>
              <div class="w-16 h-1 bg-black/8 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all" :class="scoreBarColor(stock.score.totalScore)" :style="{ width: Math.min(stock.score.totalScore, 100) + '%' }" />
              </div>
            </div>
            <StockTag v-if="stock.recommendation" type="recommendation" :text="stock.recommendation" />
          </div>
          <div class="flex items-center justify-between bg-white/70 rounded-xl px-4 py-3">
            <span class="text-xs text-text-secondary">建议仓位</span>
            <span class="text-lg font-bold num-font text-earth-sage">{{ (stock.recommendedPosition * 100).toFixed(1) }}%</span>
          </div>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <StockTag v-if="stock.macdSignal" type="macdSignal" :text="stock.macdSignal" />
            <StockTag v-if="stock.bollPosition" type="bollPosition" :text="stock.bollPosition" />
          </div>
        </div>
      </div>
    </template>

    <!-- 关注列表扫描结果 -->
    <div v-if="watchScanResult?.signals?.length" class="card p-5 space-y-3">
      <div class="text-sm font-semibold text-text-primary">关注列表扫描结果</div>
      <div v-for="stock in watchScanResult.signals" :key="stock.code"
        class="p-3.5 bg-white/70 rounded-2xl cursor-pointer transition-all hover-lift" @click="openDetail(stock.code)">
        <div class="flex justify-between items-center">
          <span class="font-bold text-text-primary">{{ stock.name }}<span class="text-text-tertiary font-normal ml-1">({{ stock.code }})</span></span>
          <span class="text-up font-medium num-font">+{{ (stock.recommendedPosition * 100).toFixed(1) }}%</span>
        </div>
        <div class="flex items-center gap-1.5 mt-1.5"><StockTag type="highlight" :text="stock.sector" /></div>
      </div>
    </div>

    <!-- 买入提醒弹窗 -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="alertVisible" class="fixed inset-0 z-50 flex items-center justify-center" @click.self="alertVisible = false">
          <div class="absolute inset-0 modal-overlay" @click="alertVisible = false" />
          <div class="card p-7 w-[90%] max-w-md shadow-2xl relative z-10">
            <div class="flex items-center gap-2 mb-5">
              <svg class="w-5 h-5 text-up" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /></svg>
              <span class="font-semibold text-text-primary">买入提醒</span>
            </div>
            <div v-if="alertStock" class="space-y-4">
              <div class="text-center">
                <div class="text-2xl font-bold text-text-primary">{{ alertStock.name }}</div>
                <div class="text-text-tertiary text-sm mt-0.5">({{ alertStock.code }})</div>
              </div>
              <div class="text-center">
                <div class="text-4xl font-bold text-up num-font">{{ (alertStock.recommendedPosition * 100).toFixed(1) }}%</div>
                <div class="text-sm text-text-tertiary">预期收益率</div>
              </div>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between p-3 bg-white/70 rounded-xl"><span class="text-text-secondary">当前价格</span><span class="font-medium text-text-primary num-font">¥{{ alertStock.price.toFixed(2) }}</span></div>
                <div class="flex justify-between p-3 bg-white/70 rounded-xl"><span class="text-text-secondary">综合评分</span><span class="font-medium num-font text-earth-sage">{{ alertStock.score }}分</span></div>
                <div class="flex justify-between p-3 bg-white/70 rounded-xl"><span class="text-text-secondary">所属板块</span><StockTag type="highlight" :text="alertStock.sector" /></div>
                <div class="flex justify-between p-3 bg-white/70 rounded-xl"><span class="text-text-secondary">资金热度</span><StockTag type="highlight" :text="alertStock.fundHeat" /></div>
              </div>
            </div>
            <button class="w-full mt-5 btn btn-primary py-3 text-sm font-medium" @click="alertVisible = false">我知道了</button>
          </div>
        </div>
      </Transition>
    </Teleport>

    <StockDetailSheet v-model:open="detailOpen" :code="selectedCode" />
  </div>
</template>
