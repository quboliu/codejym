<template>
  <div class="landing">
    <!-- Subtle Ambient Glow Overlay -->
    <div class="ambient-glow" aria-hidden="true"></div>

    <!-- Sophisticated Header -->
    <header class="lp-nav">
      <div class="lp-nav-container">
        <div class="lp-nav-left">
          <BrandLogo :size="24" wordmark-size="1.2rem" />
        </div>

        <div class="lp-nav-right">
          <button class="nav-text-btn" @click="toggleLang" :title="`Switch to ${locale === 'en' ? 'zh' : 'en'}`">
            {{ locale === 'en' ? '中' : 'En' }}
          </button>
          <button class="btn-icon theme-toggle" @click="toggleTheme" :title="`Theme: ${theme}`" aria-label="Toggle Theme" style="margin-right: var(--space-xs);">
            <svg v-if="isDark" width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M10 3V2M10 18V17M17 10H18M2 10H3M15.5 4.5L16.2 3.8M3.8 16.2L4.5 15.5M15.5 15.5L16.2 16.2M3.8 3.8L4.5 4.5" stroke-linecap="round"/>
              <circle cx="10" cy="10" r="4"/>
            </svg>
            <svg v-else width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M17 10.5C16.1 13.9 12.9 16.4 9.2 16.4C5.5 16.4 2.5 13.4 2.5 9.7C2.5 6 5 2.8 8.4 1.9C8.1 2.6 8 3.3 8 4C8 7.3 10.7 10 14 10C14.7 10 15.4 9.9 16.1 9.6L17 10.5Z" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
          <div class="nav-divider"></div>
          <button class="nav-text-btn" @click="emit('login')">{{ t('landing.logIn') }}</button>
          <button class="nav-cta-btn" @click="emit('signup')">{{ t('landing.startTraining') }}</button>
        </div>
      </div>
    </header>

    <!-- Ultra-minimalist Hero Section with Calligraphy Layout -->
    <section class="lp-hero">
      <!-- Vertical Motto Banner -->
      <div class="lp-vertical-motto animate-slideUp">
        <div class="motto-text">
          {{ t('landing.precision') }}<span class="motto-seal">码</span>
        </div>
      </div>

      <div class="lp-hero-text animate-slideUp" style="animation-delay: 0.1s;">
        <h1 class="lp-title">
          <template v-if="t('landing.masterArt')">{{ t('landing.masterArt') }}<br /></template>
          <span class="font-calligraphy text-accent" style="font-size: 1.25em;">{{ t('landing.art') }}</span> {{ t('landing.ofCode') }}
        </h1>
        <p class="lp-subtitle">
          {{ t('landing.subtitle') }}
        </p>
        <div class="lp-cta">
          <button class="btn btn-primary btn-lg" @click="emit('signup')">
            {{ t('landing.beginTrial') }}
          </button>
          <button class="btn btn-secondary btn-lg" @click="emit('login')">
            {{ t('landing.signIn') }}
          </button>
        </div>
      </div>

      <!-- Refined Monochrome Code Window -->
      <div class="lp-hero-visual animate-slideUp" style="animation-delay: 0.2s;" aria-hidden="true">
        <div class="code-window premium-border">
          <div class="cw-bar">
            <div class="cw-dots">
              <span class="cw-dot"></span>
              <span class="cw-dot"></span>
              <span class="cw-dot"></span>
            </div>
            <div class="cw-file">src/algorithms/quicksort.go</div>
            <div class="cw-spacer"></div>
          </div>
          <div class="cw-body-wrapper">
            <pre class="cw-body"><code><span class="line"><span class="dim"><span class="hl-keyword">func</span> <span class="hl-function">quickSort</span>(arr []<span class="hl-type">int</span>) []<span class="hl-type">int</span> {</span></span>
<span class="line"><span class="dim">	<span class="hl-keyword">if</span> <span class="hl-function">len</span>(arr) &lt;= <span class="hl-number">1</span> {</span></span>
<span class="line"><span class="dim">		<span class="hl-keyword">return</span> arr</span></span>
<span class="line typing">	}<span class="caret"></span></span>
<span class="line ghost">	pivot := arr[<span class="hl-function">len</span>(arr)/<span class="hl-number">2</span>]</span>
<span class="line ghost">	<span class="hl-keyword">var</span> left, right []<span class="hl-type">int</span></span>
<span class="line ghost">	</span>
<span class="line ghost">	<span class="hl-keyword">return</span> <span class="hl-function">append</span>(left, right...)</span>
<span class="line ghost">}</span></code></pre>
          </div>
        </div>
      </div>
    </section>

    <div class="divider-line"></div>

    <!-- Elegantly spaced features -->
    <section class="lp-section">
      <div class="lp-section-head animate-slideUp" style="animation-delay: 0.1s;">
        <h2>{{ t('landing.methodology') }}<span class="font-calligraphy text-accent">{{ t('landing.methodologyHighlight') }}</span></h2>
        <p class="text-secondary text-lg font-light">{{ t('landing.methodologySub') }}</p>
      </div>
      <div class="lp-features">
        <div class="lp-feature premium-border animate-slideUp" v-for="(f, idx) in featureKeys" :key="f" :style="`animation-delay: ${0.2 + idx * 0.1}s;`">
          <div class="lp-feature-icon" v-html="featureIcons[f]"></div>
          <h3>{{ t(`landing.features.${f}.title`) }}</h3>
          <p class="text-secondary font-light text-sm">{{ t(`landing.features.${f}.desc`) }}</p>
        </div>
      </div>
    </section>

    <div class="divider-line"></div>

    <!-- Minimalist Call to Action -->
    <section class="lp-final">
      <div class="lp-final-content animate-slideUp">
        <h2>{{ t('landing.elevate') }}<span class="font-calligraphy text-accent">{{ t('landing.craft') }}</span>.</h2>
        <p class="text-secondary text-lg mb-xl font-light">{{ t('landing.elevateSub') }}</p>
        <button class="btn btn-primary btn-lg" @click="emit('signup')">
          {{ t('landing.startSession') }}
        </button>
      </div>
    </section>

    <!-- Clean Footer -->
    <footer class="lp-footer">
      <div class="footer-left">
        <BrandLogo :size="16" wordmark-size="0.9rem" />
      </div>
      <div class="footer-right">
        <span class="text-tertiary text-xs">{{ t('landing.rights') }}</span>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BrandLogo from './BrandLogo.vue'
import { useTheme } from '../composables/useTheme'

const emit = defineEmits<{
  (e: 'login'): void
  (e: 'signup'): void
}>()

const { t, locale } = useI18n()
const { theme, isDark, toggleTheme } = useTheme()

function toggleLang() {
  locale.value = locale.value === 'en' ? 'zh' : 'en'
}

const featureKeys = ['tracing', 'rendering', 'fidelity', 'state', 'sync', 'import']

const featureIcons: Record<string, string> = {
  tracing: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><path d="M12 19l7-7 3 3-7 7-3-3z"/><path d="M18 13l-1.5-7.5L2 2l3.5 14.5L13 18l5-5z"/><path d="M2 2l7.586 7.586"/><circle cx="11" cy="11" r="2"/></svg>',
  rendering: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>',
  fidelity: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>',
  state: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/><polyline points="17 21 17 13 7 13 7 21"/><polyline points="7 3 7 8 15 8"/></svg>',
  sync: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>',
  import: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="square"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>',
}
</script>

<style scoped>
.landing {
  min-height: 100vh;
  position: relative;
  overflow-x: hidden;
  background: var(--color-bg-primary);
}

.ambient-glow {
  position: absolute;
  top: 0;
  right: 0;
  width: 800px;
  height: 800px;
  background: radial-gradient(circle, var(--color-glow, var(--color-gold-muted)) 0%, transparent 60%);
  transform: translate(30%, -30%);
  z-index: 0;
  pointer-events: none;
}

.divider-line {
  width: 100%;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--color-border), transparent);
}

/* ---------- Header ---------- */
.lp-nav {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 100;
  height: 80px;
  background: var(--color-bg-elevated);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--color-border);
  transition: all var(--transition-base);
}

.lp-nav-container {
  max-width: var(--max-content-width);
  margin: 0 auto;
  height: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 var(--space-2xl);
}

.lp-nav-left {
  display: flex;
  align-items: center;
}

.nav-link {
  font-family: var(--font-display);
  font-size: 0.95rem;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: color var(--transition-fast);
  letter-spacing: 0.05em;
}
.nav-link:hover {
  color: var(--color-text-primary);
}

.lp-nav-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-md);
}

.nav-divider {
  width: 1px;
  height: 16px;
  background: var(--color-border);
  margin: 0 var(--space-xs);
}

.nav-text-btn {
  background: none;
  border: none;
  font-family: var(--font-sans);
  font-size: 0.9rem;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color var(--transition-fast);
  letter-spacing: 0.02em;
}
.nav-text-btn:hover {
  color: var(--color-text-primary);
}

.nav-cta-btn {
  background: var(--color-text-primary);
  color: var(--color-bg-primary);
  border: 1px solid var(--color-text-primary);
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-family: var(--font-sans);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}
.nav-cta-btn:hover {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

/* ---------- Hero Section ---------- */
.lp-hero {
  position: relative;
  z-index: 10;
  display: grid;
  grid-template-columns: auto 1fr 1fr;
  gap: var(--space-4xl);
  align-items: center;
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: calc(var(--header-height) + var(--space-3xl)) var(--space-2xl) var(--space-3xl);
  min-height: 90vh;
}

/* Calligraphy Vertical Motto */
.lp-vertical-motto {
  writing-mode: vertical-rl;
  text-orientation: mixed;
  display: flex;
  align-items: center;
  border-left: 1px solid var(--color-border);
  padding-left: var(--space-xl);
  height: 100%;
}

.motto-text {
  font-family: var(--font-calligraphy);
  font-size: var(--font-size-2xl);
  color: var(--color-text-secondary);
  letter-spacing: 0.15em;
  white-space: nowrap; /* Prevent wrapping so seal is strictly at the end */
}

.motto-seal {
  width: 32px;
  height: 32px;
  background: var(--color-accent);
  color: var(--color-bg-primary);
  font-family: var(--font-calligraphy);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-size: 1.1rem;
  writing-mode: horizontal-tb;
  transform: rotate(-3deg);
  box-shadow: 2px 2px 8px rgba(139, 43, 34, 0.2);
  margin-top: var(--space-sm);
}

.lp-hero-text {
  max-width: 500px;
}

.lp-title {
  font-size: var(--font-size-4xl);
  line-height: 1.05;
  margin-bottom: var(--space-lg);
}

.lp-subtitle {
  font-size: var(--font-size-lg);
  color: var(--color-text-secondary);
  font-weight: 300;
  margin-bottom: var(--space-2xl);
  line-height: 1.5;
}

.lp-cta {
  display: flex;
  gap: var(--space-md);
}

/* ---------- Code Window ---------- */
.lp-hero-visual {
  display: flex;
  justify-content: flex-end;
  position: relative;
  perspective: 1000px;
}

.code-window {
  width: 100%;
  max-width: 500px;
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xl), var(--shadow-glow);
  overflow: hidden;
  transform: rotateY(-5deg) rotateX(2deg);
  transition: transform var(--transition-slow);
}

.code-window:hover {
  transform: rotateY(0deg) rotateX(0deg);
}

.cw-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
}

.cw-file {
  font-family: var(--font-mono);
  font-size: 0.65rem;
  color: var(--color-text-tertiary);
  text-transform: lowercase;
  letter-spacing: 0.05em;
  opacity: 0.7;
}

.cw-dots {
  display: flex;
  gap: 6px;
  width: 50px;
}

.cw-spacer {
  width: 50px;
}

.cw-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 1px solid var(--color-border-hover);
  background: transparent;
}

.cw-body-wrapper {
  padding: var(--space-xl) var(--space-2xl);
  background: linear-gradient(180deg, transparent, var(--color-bg-secondary));
}

.cw-body {
  margin: 0;
  font-family: var(--font-mono);
  font-size: 0.85rem;
  line-height: 1.45;
  color: var(--color-text-primary);
}

.cw-body .line { display: block; }
.cw-body .dim { color: var(--color-text-tertiary); }
.cw-body .typing { color: var(--color-text-primary); text-shadow: 0 0 8px rgba(255,255,255,0.2); }
.cw-body .ghost { color: var(--color-text-tertiary); opacity: 0.2; }

/* Syntax Highlighting Fake Classes */
.hl-keyword { color: #d73a49; }
.hl-function { color: #6f42c1; }
.hl-type { color: #005cc5; }
.hl-number { color: #005cc5; }

:global([data-theme="dark"]) .hl-keyword { color: #ff7b72; }
:global([data-theme="dark"]) .hl-function { color: #d2a8ff; }
:global([data-theme="dark"]) .hl-type { color: #79c0ff; }
:global([data-theme="dark"]) .hl-number { color: #79c0ff; }


.caret {
  display: inline-block;
  width: 1.5px;
  height: 1.2em;
  margin-left: 1px;
  vertical-align: middle;
  background: var(--color-text-primary);
  animation: pulse 1s step-end infinite;
  box-shadow: 0 0 4px var(--color-text-primary);
}

@keyframes pulse { 50% { opacity: 0; } }

/* ---------- Sections ---------- */
.lp-section {
  position: relative;
  z-index: 10;
  max-width: var(--max-content-width);
  margin: 0 auto;
  padding: var(--space-4xl) var(--space-2xl);
}

.lp-section-head {
  text-align: center;
  margin-bottom: var(--space-3xl);
}

.lp-section-head h2 {
  font-size: var(--font-size-3xl);
  margin-bottom: var(--space-sm);
}

/* ---------- Features ---------- */
.lp-features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-2xl);
}

.lp-feature {
  display: flex;
  flex-direction: column;
  padding: var(--space-2xl);
  background: var(--color-bg-primary); /* Uses the warm rice paper color */
  border-radius: var(--radius-lg);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.02), inset 0 0 0 1px var(--color-border);
  transition: transform var(--transition-slow), box-shadow var(--transition-slow);
}

.lp-feature:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(139, 43, 34, 0.04), inset 0 0 0 1px var(--color-border-hover);
}

.lp-feature-icon {
  margin-bottom: var(--space-lg);
  color: var(--color-accent); /* Seal red icon */
  padding: 10px;
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  align-self: flex-start;
  border: 1px solid var(--color-border);
}

.lp-feature-icon :deep(svg) { width: 20px; height: 20px; }
.lp-feature h3 { font-size: var(--font-size-lg); margin-bottom: var(--space-xs); font-family: var(--font-sans); font-weight: 500;}

/* ---------- Final CTA ---------- */
.lp-final {
  position: relative;
  z-index: 10;
  padding: var(--space-4xl) var(--space-2xl);
  text-align: center;
}

.lp-final-content {
  max-width: 600px;
  margin: 0 auto;
}

.lp-final h2 {
  font-size: var(--font-size-3xl);
  margin-bottom: var(--space-sm);
}

/* ---------- Footer ---------- */
.lp-footer {
  position: relative;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-xl) var(--space-2xl);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}

/* ---------- Responsive ---------- */
@media (max-width: 900px) {
  .lp-hero {
    grid-template-columns: 1fr;
    text-align: center;
    padding-top: calc(80px + var(--space-2xl));
    min-height: auto;
  }
  .lp-vertical-motto { display: none; }
  .lp-hero-text { margin: 0 auto; }
  .lp-hero-visual { justify-content: center; }
  .code-window { transform: none; }
  .code-window:hover { transform: none; }
  .lp-cta { justify-content: center; }
  .lp-features { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 600px) {
  .lp-features { grid-template-columns: 1fr; gap: var(--space-lg); }
  .lp-nav-container { padding: 0 var(--space-md); }
  .nav-text-btn { display: none; }
  .nav-divider { display: none; }
  .lp-footer { flex-direction: column; gap: var(--space-md); text-align: center; }
}
</style>
