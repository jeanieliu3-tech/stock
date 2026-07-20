<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { routes } from '@/router'
import { diagnoseStock, scanCoreSatellite, getWatchList, addWatchList, removeWatchList, scanWatchList } from '@/api'
import type { ScanResult, DiagnoseResult } from '@/types/stock'
import StockDetailSheet from '@/components/StockDetailSheet.vue'
import StockTag from '@/components/StockTag.vue'
import { getRecommendationStyle } from '@/utils/tagColors'

const route = useRoute()
const router = useRouter()
const rightPanelOpen = ref(false)
const mobileMenuOpen = ref(false)

// ─── Watchlist ───
const watchlist = ref<string[]>([])
const watchCode = ref('')
const watchScanLoading = ref(false)
const loadWatchlist = async () => {
  try { const r = await getWatchList(); if (r.code === 200) watchlist.value = r.data as string[] }
  catch {}
}
const addWatch = async () => {
  if (!watchCode.value.trim()) return
  try { const r = await addWatchList(watchCode.value.trim()); if (r.code === 200) { watchlist.value.push(watchCode.value.trim()); watchCode.value = '' } }
  catch (e) { console.error(e) }
}
const removeWatch = async (code: string) => {
  try { await removeWatchList(code); watchlist.value = watchlist.value.filter(c => c !== code) }
  catch (e) { console.error(e) }
}
const scanWatch = async () => {
  watchScanLoading.value = true
  try { const r = await scanWatchList(); if (r.code === 200 && (r.data as any)?.signals?.length) { alert('关注列表有新的信号！'); rightPanelOpen.value = false } }
  catch (e) { console.error(e) } finally { watchScanLoading.value = false }
}

// ─── Diagnose ───
const diagCode = ref('')
const diagnosing = ref(false)
const diagResult = ref<DiagnoseResult | null>(null)
const handleDiagnose = async () => {
  if (!diagCode.value.trim()) return
  diagnosing.value = true; diagResult.value = null
  try { const r = await diagnoseStock(diagCode.value.trim()); if (r.code === 200) diagResult.value = r.data as DiagnoseResult }
  catch {} finally { diagnosing.value = false }
}

// ─── Full Scan ───
const scanLoading = ref(false)
const scanResult = ref<ScanResult | null>(null)
const scanError = ref<string | null>(null)
const detailOpen = ref(false)
const selectedCode = ref('')
const handleFullScan = async () => {
  scanLoading.value = true; scanResult.value = null; scanError.value = null
  try {
    const r = await scanCoreSatellite()
    if (r.code === 200) {
      const d = r.data as ScanResult
      scanResult.value = d
      if (!d?.core?.length && !d?.satellite?.length) scanError.value = 'empty'
      rightPanelOpen.value = true
    } else { scanError.value = 'error'; rightPanelOpen.value = true }
  } catch { scanError.value = 'error'; rightPanelOpen.value = true }
  finally { scanLoading.value = false }
}
const openDetail = (code: string) => { selectedCode.value = code; detailOpen.value = true }

// ─── Nav ───
const navItems = routes.filter(r => r.path !== '/settings').map(r => ({
  path: r.path,
  name: r.name as string,
  title: (r.meta?.title as string) || r.name,
  icon: (r.meta?.icon as string) || 'LayoutDashboard',
}))

const iconMap: Record<string, string> = {
  home: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`,
  rank: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>`,
  position: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"/><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/></svg>`,
  strategies: `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>`,
}

const mobileNav = computed(() => [
  { path: '/', title: '首页', icon: 'home' },
  { path: '/rank', title: '排名', icon: 'rank' },
  { path: '/position', title: '持仓', icon: 'position' },
  { path: '/strategies', title: '策略', icon: 'strategies' },
])

const pageTitle = computed(() => {
  const m = routes.find(r => r.path === route.path)
  return (m?.meta?.title as string) || '首页'
})

// ─── Diagnose rec text ───
const getDiagRecText = (rec: string): string => {
  switch (rec) { case 'buy': return '建议买入'; case 'avoid': return '建议规避'; default: return '建议观望' }
}

onMounted(loadWatchlist)
</script>

<template>
  <div class="min-h-screen flex flex-col bg-page relative">
    <!-- 背景 blob -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden" aria-hidden="true">
      <div class="absolute -top-32 -left-32 w-[500px] h-[500px] opacity-[0.08]" style="background: radial-gradient(circle at 40% 50%, #8DB8A0, transparent 70%); border-radius: 50%; filter: blur(60px);" />
      <div class="absolute top-1/3 -right-48 w-[500px] h-[500px] opacity-[0.06]" style="background: radial-gradient(circle at 60% 50%, #D4A59A, transparent 70%); border-radius: 50%; filter: blur(60px);" />
      <div class="absolute bottom-0 left-1/3 w-[400px] h-[400px] opacity-[0.05]" style="background: radial-gradient(circle at 50% 50%, #A8C5D6, transparent 70%); border-radius: 50%; filter: blur(60px);" />
    </div>

    <!-- ═══ Desktop Top Navigation ═══ -->
    <header class="hidden md:flex sticky top-0 z-50 glass items-center h-16 px-6 gap-2" style="border-bottom: 1px solid rgba(255,255,255,0.25);">
      <div class="flex items-center gap-2.5 mr-6">
        <span class="w-2.5 h-2.5 rounded-full bg-earth-sage breathe-dot" />
        <span class="text-base font-semibold text-text-primary tracking-tight" style="font-family:'Outfit',sans-serif;">波段共振</span>
      </div>
      <nav class="flex items-center gap-1">
        <button v-for="item in navItems" :key="item.path"
          @click="router.push(item.path)"
          class="flex items-center gap-2 px-4 py-2 rounded-2xl text-sm font-medium transition-all duration-300"
          :class="route.path === item.path ? 'nav-active' : 'text-text-secondary hover:text-text-primary hover:bg-white/40'">
          <span :class="route.path === item.path ? 'text-earth-sage' : 'text-text-tertiary'" v-html="iconMap[item.icon.toLowerCase()] || iconMap.home" />
          {{ item.title }}
        </button>
      </nav>
      <div class="flex-1" />
      <!-- 全池扫描闪电按钮 -->
      <button class="flex items-center gap-1.5 px-3.5 py-2 rounded-2xl text-sm transition-all duration-300"
        :class="scanLoading ? 'btn-primary' : 'text-text-secondary hover:bg-white/40 hover:text-text-primary'"
        :disabled="scanLoading" @click="handleFullScan">
        <svg v-if="scanLoading" class="animate-spin" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><circle cx="12" cy="12" r="10" stroke-dasharray="30 70" stroke-linecap="round"/></svg>
        <svg v-else width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
        <span class="hidden lg:inline">{{ scanLoading ? '扫描中' : '扫描' }}</span>
      </button>
      <!-- 右侧面板切换 -->
      <button class="flex items-center gap-2 px-3.5 py-2 rounded-2xl text-sm text-text-secondary hover:bg-white/40 hover:text-text-primary transition-all duration-300" @click="rightPanelOpen = !rightPanelOpen">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>
        <span class="hidden lg:inline">工具</span>
      </button>
    </header>

    <!-- ═══ Mobile Header ═══ -->
    <header class="md:hidden sticky top-0 z-50 glass px-5 h-14 flex items-center justify-between" style="border-bottom: 1px solid rgba(255,255,255,0.25);">
      <div class="flex items-center gap-2.5">
        <span class="w-2.5 h-2.5 rounded-full bg-earth-sage breathe-dot" />
        <span class="text-base font-semibold text-text-primary tracking-tight">{{ pageTitle }}</span>
      </div>
      <div class="flex items-center gap-2">
        <button class="w-9 h-9 flex items-center justify-center rounded-2xl hover:bg-white/40 transition-all" @click="handleFullScan">
          <svg v-if="scanLoading" class="animate-spin" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2"><circle cx="12" cy="12" r="10" stroke-dasharray="30 70" stroke-linecap="round"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" class="text-text-secondary"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
        </button>
        <button class="w-9 h-9 flex items-center justify-center rounded-2xl hover:bg-white/40 transition-all" @click="rightPanelOpen = !rightPanelOpen">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" class="text-text-secondary"><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
        <button class="w-9 h-9 flex items-center justify-center rounded-2xl hover:bg-white/40 transition-all" @click="mobileMenuOpen = !mobileMenuOpen">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" class="text-text-secondary"><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        </button>
      </div>
    </header>

    <!-- ═══ Main Content ═══ -->
    <main class="flex-1 w-full mx-auto px-4 md:px-8 lg:px-12 py-5 md:py-8 overflow-auto relative z-10" style="max-width:1400px;">
      <slot />
    </main>

    <!-- ═══ Right Panel（工具面板） ═══ -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="rightPanelOpen" class="fixed inset-0 z-50 flex justify-end" @click.self="rightPanelOpen = false">
          <div class="absolute inset-0 modal-overlay" @click="rightPanelOpen = false" />
          <div class="relative w-[360px] max-w-[92vw] h-full overflow-y-auto p-5 space-y-5 glass shadow-2xl" style="border-left: 1px solid rgba(255,255,255,0.25);">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold text-text-primary">工具</h3>
              <button class="w-8 h-8 flex items-center justify-center rounded-2xl hover:bg-white/40 transition-all" @click="rightPanelOpen = false">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" class="text-text-secondary"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>

            <!-- 个股诊断 -->
            <div>
              <h4 class="text-xs font-semibold text-text-secondary mb-3">个股诊断</h4>
              <div class="flex gap-2 mb-2">
                <input v-model="diagCode" placeholder="输入股票代码" class="input-organic flex-1 text-sm" @keyup.enter="handleDiagnose" />
                <button class="btn btn-primary text-sm px-4 py-2" :class="{ 'opacity-50': diagnosing }" :disabled="diagnosing" @click="handleDiagnose">{{ diagnosing ? '...' : '诊断' }}</button>
              </div>
              <Transition name="page">
                <div v-if="diagResult" class="bg-white/70 rounded-2xl p-3.5 space-y-1.5">
                  <div class="flex justify-between items-center">
                    <span class="font-semibold text-sm text-text-primary">{{ diagResult.name }}</span>
                    <span class="text-[11px] font-medium" :class="getRecommendationStyle(getDiagRecText(diagResult.recommendation))">{{ getDiagRecText(diagResult.recommendation) }}</span>
                  </div>
                  <div class="text-xs text-text-secondary leading-relaxed">{{ diagResult.analysis }}</div>
                </div>
              </Transition>
            </div>

            <!-- 关注列表 -->
            <div>
              <h4 class="text-xs font-semibold text-text-secondary mb-3">关注列表</h4>
              <div class="flex gap-2 mb-2">
                <input v-model="watchCode" placeholder="输入代码" class="input-organic flex-1 text-sm" @keyup.enter="addWatch" />
                <button class="btn btn-primary text-sm px-4 py-2" @click="addWatch">添加</button>
              </div>
              <div v-if="watchlist.length > 0" class="space-y-1 max-h-36 overflow-y-auto">
                <div v-for="code in watchlist" :key="code" class="flex items-center justify-between px-3 py-2 rounded-2xl bg-white/40 text-xs">
                  <span class="text-text-secondary font-medium">{{ code }}</span>
                  <button class="text-text-tertiary hover:text-up transition-all" @click="removeWatch(code)">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </div>
              </div>
              <div v-else class="text-xs text-text-tertiary py-2 text-center">暂无关注</div>
              <button v-if="watchlist.length > 0" class="w-full mt-2 btn-ghost text-xs py-2 rounded-2xl" :class="{ 'opacity-50': watchScanLoading }" :disabled="watchScanLoading" @click="scanWatch">{{ watchScanLoading ? '扫描中...' : '扫描关注列表' }}</button>
            </div>

            <!-- 市场判断 -->
            <div>
              <h4 class="text-xs font-semibold text-text-secondary mb-3">市场判断</h4>
              <div class="bg-white/60 rounded-2xl p-4 space-y-3">
                <div class="relative h-2 rounded-full overflow-hidden bg-black/5 flex">
                  <div class="h-full rounded-l-full bg-earth-sage" style="width:38%" />
                  <div class="h-full bg-earth-clay-light" style="width:24%" />
                  <div class="h-full rounded-r-full bg-earth-terracotta" style="width:38%" />
                </div>
                <div class="flex justify-between text-xs">
                  <span class="text-earth-sage font-medium">偏多 38%</span>
                  <span class="text-text-tertiary">中性 24%</span>
                  <span class="text-earth-terracotta font-medium">偏空 38%</span>
                </div>
                <div class="bg-white/80 rounded-xl p-3 text-xs text-text-secondary">
                  参考仓位 <span class="text-earth-sage font-semibold">&lt; 30%</span>
                  <span class="block text-[10px] text-text-tertiary mt-1">基于历史波动率估算</span>
                </div>
              </div>
            </div>

            <!-- 扫描结果 -->
            <div v-if="scanResult || scanError">
              <h4 class="text-xs font-semibold text-text-secondary mb-3">扫描结果</h4>
              <div v-if="scanError === 'empty'" class="bg-black/4 rounded-2xl px-4 py-3 text-xs text-text-tertiary text-center">当前策略未匹配到股票</div>
              <div v-else-if="scanError === 'error'" class="bg-[#F5E3E0]/60 rounded-2xl px-4 py-3 text-xs text-[#A86B62] text-center">扫描服务异常，请稍后重试</div>
              <div v-else-if="scanResult" class="space-y-2">
                <div v-if="scanResult.core?.length" class="space-y-1.5">
                  <div class="flex items-center gap-1.5"><div class="w-1 h-4 rounded-full bg-earth-sage" /><span class="text-xs font-semibold text-earth-sage">核心 ({{ scanResult.core.length }})</span></div>
                  <div v-for="s in scanResult.core.slice(0,5)" :key="s.code" class="bg-white/60 rounded-2xl px-3.5 py-2.5 cursor-pointer transition-all hover:bg-white/80" @click="openDetail(s.code); rightPanelOpen = false">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-medium text-text-primary">{{ s.name }}</span>
                      <span class="text-xs font-semibold num-font" :class="s.changePercent >= 0 ? 'text-up' : 'text-down'">{{ s.changePercent >= 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%</span>
                    </div>
                    <div class="flex items-center gap-2 mt-0.5">
                      <span class="text-[10px] text-text-tertiary">{{ s.code }}</span>
                      <StockTag v-if="s.recommendation" type="recommendation" :text="s.recommendation" />
                    </div>
                  </div>
                  <div v-if="scanResult.core.length > 5" class="text-[10px] text-text-tertiary text-center pt-1">+{{ scanResult.core.length - 5 }} 更多</div>
                </div>
                <div v-if="scanResult.satellite?.length" class="space-y-1.5">
                  <div class="flex items-center gap-1.5 mt-3"><div class="w-1 h-4 rounded-full bg-earth-sky" /><span class="text-xs font-semibold text-earth-sky">卫星 ({{ scanResult.satellite.length }})</span></div>
                  <div v-for="s in scanResult.satellite.slice(0,3)" :key="s.code" class="bg-white/60 rounded-2xl px-3.5 py-2.5 cursor-pointer transition-all hover:bg-white/80" @click="openDetail(s.code); rightPanelOpen = false">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-medium text-text-primary">{{ s.name }}</span>
                      <span class="text-xs font-semibold num-font" :class="s.changePercent >= 0 ? 'text-up' : 'text-down'">{{ s.changePercent >= 0 ? '+' : '' }}{{ s.changePercent.toFixed(2) }}%</span>
                    </div>
                    <div class="text-[10px] text-text-tertiary mt-0.5">{{ s.code }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ═══ Mobile Bottom Nav ═══ -->
    <nav class="md:hidden fixed bottom-0 left-0 right-0 glass z-40" style="border-top: 1px solid rgba(255,255,255,0.25);">
      <div class="flex items-center justify-around h-16 px-2">
        <button v-for="item in mobileNav" :key="item.path"
          @click="router.push(item.path)"
          class="flex flex-col items-center justify-center gap-0.5 flex-1 py-1 transition-all duration-300 rounded-2xl"
          :class="route.path === item.path ? 'text-earth-sage bg-white/40' : 'text-text-tertiary hover:text-text-secondary'">
          <span v-html="iconMap[item.icon]" />
          <span class="text-[10px] font-medium">{{ item.title }}</span>
        </button>
      </div>
    </nav>
    <div class="md:hidden h-16" />

    <!-- Mobile Menu -->
    <Teleport to="body">
      <Transition name="drawer">
        <div v-if="mobileMenuOpen" class="fixed inset-0 z-50 flex md:hidden" @click.self="mobileMenuOpen = false">
          <div class="absolute inset-0 modal-overlay" @click="mobileMenuOpen = false" />
          <div class="relative ml-auto w-[260px] h-full overflow-y-auto p-5 space-y-4 glass shadow-2xl" style="border-left: 1px solid rgba(255,255,255,0.25);">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-semibold text-text-primary">导航</span>
              <button class="w-8 h-8 flex items-center justify-center rounded-2xl hover:bg-white/40 transition-all" @click="mobileMenuOpen = false">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" class="text-text-secondary"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <div class="space-y-1">
              <button v-for="item in navItems" :key="item.path"
                @click="router.push(item.path); mobileMenuOpen = false"
                class="w-full flex items-center gap-3 px-3.5 py-3 rounded-2xl text-sm transition-all duration-300"
                :class="route.path === item.path ? 'bg-white/70 shadow-sm font-medium nav-active' : 'text-text-secondary hover:bg-white/40 hover:text-text-primary'">
                <span :class="route.path === item.path ? 'text-earth-sage' : 'text-text-tertiary'" v-html="iconMap[item.icon.toLowerCase()] || iconMap.home" />
                {{ item.title }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- StockDetailSheet -->
    <StockDetailSheet :open="detailOpen" :code="selectedCode" @update:open="detailOpen = $event" />
  </div>
</template>
