<template>
  <div class="landing">
    <!-- 顶部导航 -->
    <header class="lp-nav">
      <BrandLogo :size="26" wordmark-size="1.15rem" />
      <nav class="lp-nav-actions">
        <button class="btn-icon" @click="toggleTheme" :title="`当前主题: ${theme}`" aria-label="切换主题">
          <svg v-if="isDark" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 3V2M10 18V17M17 10H18M2 10H3M15.5 4.5L16.2 3.8M3.8 16.2L4.5 15.5M15.5 15.5L16.2 16.2M3.8 3.8L4.5 4.5" stroke-linecap="round"/>
            <circle cx="10" cy="10" r="4"/>
          </svg>
          <svg v-else width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 10.5C16.1 13.9 12.9 16.4 9.2 16.4C5.5 16.4 2.5 13.4 2.5 9.7C2.5 6 5 2.8 8.4 1.9C8.1 2.6 8 3.3 8 4C8 7.3 10.7 10 14 10C14.7 10 15.4 9.9 16.1 9.6L17 10.5Z" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
        <button class="btn btn-ghost" @click="emit('login')">登录</button>
        <button class="btn btn-primary" @click="emit('signup')">免费开始</button>
      </nav>
    </header>

    <!-- 英雄区 -->
    <section class="lp-hero">
      <div class="lp-hero-text">
        <span class="lp-eyebrow">代码临摹 · 刻意练习</span>
        <h1 class="lp-title">把好代码<br />练成肌肉记忆</h1>
        <p class="lp-subtitle">
          CodeJYM 是你的「代码健身房」。导入优秀的开源项目，逐字临摹，
          在重复中内化高手的写法 —— 进度自动保存，随时随地接着练。
        </p>
        <div class="lp-cta">
          <button class="btn btn-primary btn-lg" @click="emit('signup')">免费开始训练</button>
          <button class="btn btn-secondary btn-lg" @click="emit('login')">我已有账号</button>
        </div>
        <div class="lp-stats">
          <div class="lp-stat"><strong>20+</strong><span>支持语言</span></div>
          <div class="lp-stat-divider"></div>
          <div class="lp-stat"><strong>自动</strong><span>进度保存</span></div>
          <div class="lp-stat-divider"></div>
          <div class="lp-stat"><strong>云端</strong><span>跨设备同步</span></div>
        </div>
      </div>

      <!-- 临摹效果预览 -->
      <div class="lp-hero-visual" aria-hidden="true">
        <div class="code-window">
          <div class="cw-bar">
            <span class="cw-dot"></span><span class="cw-dot"></span><span class="cw-dot"></span>
            <span class="cw-file">quicksort.go</span>
          </div>
          <pre class="cw-body"><code><span class="done">func quickSort(arr []int) []int {</span>
<span class="done">	if len(arr) &lt;= 1 {</span>
<span class="done">		return arr</span>
<span class="typing">	}<span class="caret"></span></span>
<span class="ghost">	pivot := arr[len(arr)/2]</span>
<span class="ghost">	var left, right []int</span>
<span class="ghost">	// 临摹这一行，肌肉记忆就+1</span>
<span class="ghost">	return append(left, right...)</span>
<span class="ghost">}</span></code></pre>
        </div>
      </div>
    </section>

    <!-- 特性 -->
    <section class="lp-section">
      <div class="lp-section-head">
        <h2>为什么用临摹来学代码？</h2>
        <p class="text-secondary">读懂 ≠ 写得出。临摹让你的手跟上你的眼。</p>
      </div>
      <div class="lp-features">
        <div class="lp-feature card" v-for="f in features" :key="f.title">
          <div class="lp-feature-icon" v-html="f.icon"></div>
          <h3>{{ f.title }}</h3>
          <p class="text-secondary">{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <!-- 三步开始 -->
    <section class="lp-section lp-steps-section">
      <div class="lp-section-head">
        <h2>三步开始训练</h2>
      </div>
      <div class="lp-steps">
        <div class="lp-step" v-for="(s, i) in steps" :key="s.title">
          <div class="lp-step-num">{{ i + 1 }}</div>
          <h3>{{ s.title }}</h3>
          <p class="text-secondary">{{ s.desc }}</p>
        </div>
      </div>
    </section>

    <!-- 结尾 CTA -->
    <section class="lp-final">
      <BrandLogo :size="40" :with-wordmark="false" badge />
      <h2>准备好开始你的第一组训练了吗？</h2>
      <p class="text-secondary">免费注册，导入第一份代码，现在就开始临摹。</p>
      <button class="btn btn-primary btn-lg" @click="emit('signup')">免费开始训练</button>
    </section>

    <!-- 页脚 -->
    <footer class="lp-footer">
      <BrandLogo :size="20" wordmark-size="1rem" />
      <span class="text-tertiary">代码临摹，刻意练习 · © 2026 CodeJYM</span>
    </footer>
  </div>
</template>

<script setup lang="ts">
import BrandLogo from './BrandLogo.vue'
import { useTheme } from '../composables/useTheme'

const emit = defineEmits<{
  (e: 'login'): void
  (e: 'signup'): void
}>()

const { theme, isDark, toggleTheme } = useTheme()

const features = [
  {
    title: '逐字临摹',
    desc: '原代码作底稿，你照着敲。敲对才前进，错了即时提示。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 19l7-7 3 3-7 7-3-3z"/><path d="M18 13l-1.5-7.5L2 2l3.5 14.5L13 18l5-5z"/><path d="M2 2l7.586 7.586"/><circle cx="11" cy="11" r="2"/></svg>',
  },
  {
    title: '语法高亮',
    desc: '20+ 编程语言的高亮渲染，临摹时上下文一目了然。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>',
  },
  {
    title: '原文逐字校对',
    desc: '注释、字符串和代码都按原文练习，敲对才继续前进。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>',
  },
  {
    title: '进度自动保存',
    desc: '光标位置、用时、错误数实时记录，关掉再回来无缝接着练。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>',
  },
  {
    title: '跨设备同步',
    desc: '数据存在云端，公司电脑练一半，回家用笔记本接着练。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>',
  },
  {
    title: '灵活导入',
    desc: '上传 ZIP 项目、单个文件，或直接粘贴代码片段，即刻开练。',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>',
  },
]

const steps = [
  { title: '导入代码', desc: '上传一份你想吃透的优秀代码 —— 开源项目、面试题解、或自己的旧作。' },
  { title: '逐行临摹', desc: '照着原文一个字符一个字符地敲，系统帮你校对每一次输入。' },
  { title: '追踪进步', desc: '每次练习都被记录，用时和准确率的提升看得见。' },
]
</script>

<style scoped>
.landing {
  min-height: 100vh;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

/* ---------- 导航 ---------- */
.lp-nav {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-md) var(--space-xl);
  background: color-mix(in srgb, var(--color-bg-primary) 80%, transparent);
  backdrop-filter: saturate(180%) blur(10px);
  border-bottom: 1px solid var(--color-border);
}
.lp-nav-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

/* ---------- 英雄区 ---------- */
.lp-hero {
  display: grid;
  grid-template-columns: 1.05fr 0.95fr;
  gap: var(--space-3xl);
  align-items: center;
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-3xl) var(--space-xl);
}
.lp-eyebrow {
  display: inline-block;
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-secondary);
  padding: var(--space-xs) var(--space-md);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-full);
  margin-bottom: var(--space-lg);
}
.lp-title {
  font-size: clamp(2.25rem, 5vw, 3.75rem);
  font-weight: 700;
  line-height: 1.08;
  letter-spacing: -0.03em;
  margin-bottom: var(--space-lg);
}
.lp-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  max-width: 34ch;
  margin-bottom: var(--space-xl);
}
.lp-cta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-md);
  margin-bottom: var(--space-2xl);
}
.lp-stats {
  display: flex;
  align-items: center;
  gap: var(--space-lg);
}
.lp-stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.lp-stat strong {
  font-size: var(--font-size-xl);
  font-weight: 700;
  letter-spacing: -0.02em;
}
.lp-stat span {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}
.lp-stat-divider {
  width: 1px;
  height: 32px;
  background: var(--color-border);
}

/* ---------- 代码窗口 ---------- */
.lp-hero-visual {
  display: flex;
  justify-content: center;
}
.code-window {
  width: 100%;
  max-width: 460px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl);
  overflow: hidden;
}
.cw-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: var(--space-sm) var(--space-md);
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
}
.cw-dot {
  width: 10px;
  height: 10px;
  border-radius: var(--radius-full);
  background: var(--color-border-hover);
}
.cw-file {
  margin-left: var(--space-sm);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}
.cw-body {
  margin: 0;
  padding: var(--space-lg);
  font-family: var(--font-mono);
  font-size: 0.8rem;
  line-height: 1.7;
  overflow-x: auto;
}
.cw-body .done { color: var(--color-text-primary); }
.cw-body .typing { color: var(--color-text-primary); }
.cw-body .ghost { color: var(--color-text-tertiary); opacity: 0.5; }
.caret {
  display: inline-block;
  width: 2px;
  height: 1em;
  margin-left: 1px;
  vertical-align: text-bottom;
  background: var(--color-accent);
  animation: cjBlink 1.1s step-end infinite;
}
@keyframes cjBlink { 50% { opacity: 0; } }

/* ---------- 通用区块 ---------- */
.lp-section {
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-3xl) var(--space-xl);
}
.lp-section-head {
  text-align: center;
  margin-bottom: var(--space-2xl);
}
.lp-section-head h2 {
  font-size: clamp(1.75rem, 3vw, 2.25rem);
  letter-spacing: -0.02em;
  margin-bottom: var(--space-sm);
}

/* ---------- 特性 ---------- */
.lp-features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-lg);
}
.lp-feature {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}
.lp-feature-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: var(--color-accent-subtle);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
  margin-bottom: var(--space-sm);
}
.lp-feature-icon :deep(svg) { width: 22px; height: 22px; }
.lp-feature h3 { font-size: var(--font-size-lg); }

/* ---------- 步骤 ---------- */
.lp-steps-section { background: var(--color-bg-secondary); max-width: none; }
.lp-steps {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-xl);
  max-width: var(--max-content-width);
  margin: 0 auto;
}
.lp-step { text-align: center; padding: 0 var(--space-md); }
.lp-step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  margin-bottom: var(--space-md);
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--color-text-inverse);
  background: var(--color-accent);
  border-radius: var(--radius-full);
}
.lp-step h3 { margin-bottom: var(--space-xs); }

/* ---------- 结尾 CTA ---------- */
.lp-final {
  text-align: center;
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-3xl) var(--space-xl);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-md);
}
.lp-final h2 {
  font-size: clamp(1.75rem, 3vw, 2.25rem);
  letter-spacing: -0.02em;
  margin-top: var(--space-sm);
}
.lp-final .btn { margin-top: var(--space-md); }

/* ---------- 页脚 ---------- */
.lp-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-md);
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-xl);
  border-top: 1px solid var(--color-border);
  font-size: var(--font-size-sm);
}

/* ---------- 响应式 ---------- */
@media (max-width: 900px) {
  .lp-hero { grid-template-columns: 1fr; gap: var(--space-2xl); }
  .lp-hero-visual { order: -1; }
  .lp-features { grid-template-columns: repeat(2, 1fr); }
  .lp-steps { grid-template-columns: 1fr; gap: var(--space-lg); }
}
@media (max-width: 560px) {
  .lp-nav { padding: var(--space-sm) var(--space-lg); }
  .lp-nav-actions .btn-ghost { display: none; }
  .lp-features { grid-template-columns: 1fr; }
  .lp-subtitle { max-width: none; }
}
</style>
