/**
 * 有机极简 — 标签色彩规范
 * 大地色系、低饱和、柔和质感
 */

export interface TagStyle {
  bg: string
  text: string
}

// 亮点标签 - 柔和底 + 深色字
export const TAG_STYLE_MAP: Record<string, TagStyle> = {
  '水上金叉':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  'MACD金叉':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '红柱放大':   { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '水下金叉':   { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
  '即将金叉':   { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
  '强势上涨':   { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '强势大涨':   { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '突破上轨':   { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
  '量能爆发':   { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '稳步上涨':   { bg: 'bg-[#E3F0EA]', text: 'text-[#4A8A6A]' },
  '明显放量':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '温和放量':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '资金爆量':   { bg: 'bg-[#EDE6F2]', text: 'text-[#7A62A0]' },
  '资金活跃':   { bg: 'bg-[#EDE6F2]', text: 'text-[#7A62A0]' },
  '资金关注':   { bg: 'bg-[#EDE6F2]', text: 'text-[#7A62A0]' },
  '超大成交':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '大额成交':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '布林开口':   { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '超卖反弹':   { bg: 'bg-[#E3F0EA]', text: 'text-[#4A8A6A]' },
  '板块龙头':   { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
}

export function getTagStyle(tag: string): TagStyle {
  return TAG_STYLE_MAP[tag] || { bg: 'bg-black/4', text: 'text-text-secondary' }
}

// 推荐等级（纯文字加粗）
export const RECOMMENDATION_STYLE_MAP: Record<string, string> = {
  '强烈推荐': 'text-earth-sage font-semibold',
  '积极关注': 'text-earth-sage font-semibold',
  '一般关注': 'text-text-secondary',
  '观望':     'text-text-tertiary',
  '谨慎观望': 'text-text-tertiary',
  '回避':     'text-text-tertiary',
}

export function getRecommendationStyle(rec: string): string {
  return RECOMMENDATION_STYLE_MAP[rec] || 'text-text-tertiary'
}

// MACD 信号标签
export const MACD_SIGNAL_STYLE_MAP: Record<string, TagStyle> = {
  '水上金叉': { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '水下金叉': { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
  '金叉':     { bg: 'bg-[#E3EDF5]', text: 'text-[#4A7A9E]' },
  '即将金叉': { bg: 'bg-[#F5EDDF]', text: 'text-[#A8884A]' },
  '死叉':     { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '弱死叉':   { bg: 'bg-[#F5E3E0]', text: 'text-[#A86B62]' },
  '待分析':   { bg: 'bg-black/4', text: 'text-[#9A9A9A]' },
}

export function getMacdSignalStyle(signal: string): TagStyle {
  return MACD_SIGNAL_STYLE_MAP[signal] || { bg: 'bg-black/4', text: 'text-[#9A9A9A]' }
}

// BOLL位置
export const BOLL_POSITION_STYLE_MAP: Record<string, string> = {
  '突破上轨': 'text-[#A86B62] font-medium',
  '上轨区域': 'text-[#A8884A]',
  '中轨上方': 'text-[#4A7A9E]',
  '中轨下方': 'text-[#6B6B6B]',
  '下轨区域': 'text-[#6B6B6B]',
  '跌破下轨': 'text-[#4A8A6A]',
  '待分析':   'text-[#9A9A9A]',
}
