<script lang="ts">
  import Terminal from '$lib/Terminal.svelte';

  const installCommand = 'go install github.com/almonk/bontree@latest';
  let copied = false;
  let copiedTimer: ReturnType<typeof setTimeout> | null = null;

  async function copyInstallCommand() {
    try {
      await navigator.clipboard.writeText(installCommand);
      copied = true;
      if (copiedTimer) clearTimeout(copiedTimer);
      copiedTimer = setTimeout(() => {
        copied = false;
        copiedTimer = null;
      }, 1200);
    } catch {
      copied = false;
    }
  }

  function onInstallKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      void copyInstallCommand();
    }
  }
</script>

<svelte:head>
  <title>bontree | Pretty + fast file tree for agent UIs</title>
  <meta name="description" content="bontree is a pretty, fast terminal file tree built to pair with agent-driven coding workflows." />
  <meta name="keywords" content="bontree,file tree,terminal ui,cli,go,coding agent,developer tools" />
  <meta name="author" content="Alasdair Monk" />
  <meta name="robots" content="index,follow" />
  <meta name="theme-color" content="#1a1b26" />
  <meta name="application-name" content="bontree" />
  <meta name="apple-mobile-web-app-title" content="bontree" />

  <meta property="og:type" content="website" />
  <meta property="og:site_name" content="bontree" />
  <meta property="og:title" content="bontree | Pretty + fast file tree for agent UIs" />
  <meta property="og:description" content="A pretty, fast terminal file tree built to pair with agent-driven coding workflows." />
  <meta property="og:url" content="https://github.com/almonk/bontree" />
  <meta property="og:image" content="/favicon.ico" />

  <meta name="twitter:card" content="summary" />
  <meta name="twitter:title" content="bontree | Pretty + fast file tree for agent UIs" />
  <meta name="twitter:description" content="A pretty, fast terminal file tree built to pair with agent-driven coding workflows." />
  <meta name="twitter:image" content="/favicon.ico" />
</svelte:head>

<div class="wrapper">
  <p class="tagline"><span class="tagline-brand">bontree</span> - a pretty + fast file tree to pair with an agent ui</p>
  <div class="terminal-slot">
    <Terminal />
  </div>
  <div class="meta-bar">
    <div class="meta-left">
      <a
        class="pill repo"
        href="https://github.com/almonk/bontree"
        target="_blank"
        rel="noopener noreferrer"
        aria-label="Open almonk/bontree on GitHub"
      >
        <svg viewBox="0 0 24 24" class="icon" fill="currentColor" aria-hidden="true">
          <path d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.009-.866-.014-1.7-2.782.604-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
        </svg>
        <span>almonk/bontree</span>
      </a>
      <div
        class="pill install install-pill"
        role="button"
        tabindex="0"
        title={installCommand}
        aria-label="Copy Go install command"
        on:click={copyInstallCommand}
        on:keydown={onInstallKeydown}
      >
        <code class="install-text" aria-label="go install github.com/almonk/bontree at latest">
          <span class="tok-cmd">go</span>
          <span class="tok-subcmd"> install </span>
          <span class="tok-path">github.com/almonk/bontree</span>
          <span class="tok-ver">@…</span>
        </code>
        <button
          class="copy-btn"
          type="button"
          aria-label="Copy Go install command"
          on:click|stopPropagation={copyInstallCommand}
        >
          {#if copied}
            <svg viewBox="0 0 24 24" class="copy-icon" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path d="M20 6L9 17l-5-5" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" class="copy-icon" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <rect x="9" y="9" width="11" height="11" rx="2" />
              <path d="M5 15V6a2 2 0 0 1 2-2h9" stroke-linecap="round" />
            </svg>
          {/if}
        </button>
      </div>
    </div>
    <a
      class="pill author"
      href="https://alasdairmonk.com"
      target="_blank"
      rel="noopener noreferrer"
      aria-label="Open almonk website"
    >
      <svg viewBox="0 0 330 330" class="icon" fill="currentColor" aria-hidden="true">
        <path d="M255.9678,275.7774 L216.8818,167.8704 L232.0048,135.2544 L277.3398,254.0264 C273.9128,258.3414 270.2478,262.4624 266.3558,266.3574 C263.0478,269.6634 259.5798,272.8054 255.9678,275.7774 L255.9678,275.7774 Z M90.9308,288.0804 L123.1818,196.8364 L153.3858,263.5484 C155.0818,267.2994 158.7878,269.6994 162.9038,269.7184 L162.9478,269.7184 C167.0468,269.7194 170.7528,267.3554 172.4778,263.6354 L204.3408,194.9164 L238.1628,288.2924 C216.6758,301.0954 191.5858,308.4504 164.7268,308.4534 C137.7198,308.4504 112.4978,301.0154 90.9308,288.0804 L90.9308,288.0804 Z M51.3438,253.0534 L95.3398,135.3404 L110.6738,169.2104 L73.1148,275.4714 C69.6368,272.5904 66.2918,269.5494 63.0978,266.3574 C58.9118,262.1704 54.9888,257.7244 51.3438,253.0534 L51.3438,253.0534 Z M136.6298,158.7884 L163.6088,82.4604 L191.2548,158.7884 L136.6298,158.7884 Z M188.2088,179.7874 L163.0648,234.0124 L138.5148,179.7874 L188.2088,179.7874 Z M63.0978,63.0974 C89.1398,37.0704 125.0038,21.0064 164.7268,21.0024 C204.4478,21.0064 240.3128,37.0704 266.3558,63.0974 C292.3798,89.1384 308.4468,125.0024 308.4508,164.7244 C308.4478,189.2114 302.3348,212.2264 291.5558,232.3834 L242.7558,104.5244 C241.2518,100.5874 237.5228,97.9204 233.3098,97.7754 C229.0988,97.6274 225.1908,100.0274 223.4208,103.8524 L206.7308,139.8454 L173.3608,47.7194 C171.8478,43.5444 167.8868,40.7774 163.4478,40.7954 C159.0078,40.8124 155.0668,43.6094 153.5888,47.7954 L120.7808,140.6164 L103.9768,103.5004 C102.2268,99.6324 98.3018,97.1944 94.0578,97.3394 C89.8148,97.4804 86.0638,100.1794 84.5768,104.1554 L37.1658,231.0014 C26.8418,211.1764 21.0038,188.6524 20.9998,164.7244 C21.0068,125.0044 37.0708,89.1384 63.0978,63.0974 L63.0978,63.0974 Z M164.7268,0.0004 C73.7468,0.0084 0.0068,73.7474 -0.0002,164.7244 C0.0068,255.7084 73.7468,329.4464 164.7268,329.4554 C255.7058,329.4484 329.4448,255.7084 329.4498,164.7244 C329.4448,73.7474 255.7058,0.0064 164.7268,0.0004 L164.7268,0.0004 Z" />
      </svg>
      <span class="author-label">made by @almonk</span>
    </a>
  </div>
</div>

<style>
  .wrapper {
    --pill-gap: 10px;
    width: 100%;
    max-width: 720px;
    padding: 24px;
  }

  .meta-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--pill-gap);
    margin-top: 14px;
    flex-wrap: nowrap;
    order: 3;
  }

  .meta-left {
    display: flex;
    align-items: center;
    gap: var(--pill-gap);
    flex-wrap: nowrap;
    min-width: 0;
    flex: 1 1 auto;
  }

  .pill {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 34px;
    padding: 0 14px 0 10px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: color-mix(in oklab, var(--bg-card) 84%, black);
    color: var(--fg);
    text-decoration: none;
    font-size: 13px;
    line-height: 1;
    white-space: nowrap;
    min-width: 0;
    transition: filter 120ms ease-out, transform 80ms ease-out;
    transform-origin: center;
  }

  .pill > span {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .pill:hover {
    filter: brightness(1.08);
  }

  .pill:active {
    transform: scale(0.985);
  }

  .install {
    font-family: inherit;
    font-size: 12px;
    color: var(--fg-dim);
  }

  .install-pill {
    cursor: pointer;
    user-select: none;
    gap: 10px;
    padding-left: 14px;
    padding-right: 6px;
    min-width: 0;
    flex: 1 1 auto;
  }

  .install-pill:focus-visible {
    outline: 1px solid var(--accent);
    outline-offset: 2px;
  }

  .install-text {
    font-family: inherit;
    color: var(--fg-dim);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tok-cmd {
    color: var(--green);
  }

  .tok-subcmd {
    color: var(--fg-dim);
  }

  .tok-path {
    color: var(--cyan);
  }

  .tok-ver {
    color: var(--yellow);
  }

  .copy-btn {
    border: 1px solid var(--border);
    background: color-mix(in oklab, var(--bg-lighter) 86%, black);
    color: var(--fg);
    border-radius: 999px;
    height: 22px;
    padding: 0 8px;
    font: inherit;
    font-size: 11px;
    line-height: 1;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
  }

  .copy-icon {
    width: 12px;
    height: 12px;
  }

  .author {
    margin-left: auto;
  }

  .icon {
    width: 16px;
    height: 16px;
    flex: 0 0 auto;
  }

  .author .author-label {
    overflow: visible;
    text-overflow: clip;
    white-space: nowrap;
  }

  .tagline {
    margin: 0 0 10px 0;
    color: var(--fg);
    font-size: 14px;
    line-height: 1.25;
    font-style: italic;
    text-wrap: balance;
    order: 1;
  }

  .terminal-slot {
    order: 2;
  }

  .tagline-brand {
    color: #fff;
    font-style: normal;
    font-size: 13px;
  }

  @media (max-width: 780px) {
    .wrapper {
      padding: 12px;
    }

    .meta-bar {
      flex-wrap: wrap;
      row-gap: var(--pill-gap);
    }

    .author {
      margin-left: 0;
      flex: 0 0 auto;
    }

    .pill {
      max-width: none;
    }

    .install {
      white-space: nowrap;
      line-height: 1;
      height: 34px;
      min-height: 34px;
      padding-top: 0;
      padding-bottom: 0;
    }

    .install-pill {
      flex: 0 1 auto;
      max-width: 100%;
    }
  }

  @media (max-width: 640px) {
    .install-pill {
      display: none;
    }

    .meta-bar {
      flex-direction: row;
      align-items: center;
      flex-wrap: nowrap;
      order: 2;
      margin-top: 10px;
      margin-bottom: 10px;
    }

    .meta-left {
      flex-wrap: nowrap;
      gap: var(--pill-gap);
      flex: 1 1 auto;
      min-width: 0;
    }

    .repo {
      flex: 0 1 auto;
      min-width: 0;
    }

    .author {
      width: fit-content;
      align-self: auto;
      flex: 0 0 auto;
    }

    .terminal-slot {
      order: 3;
    }
  }

  @media (max-width: 420px) {
    .wrapper {
      padding: 8px;
    }

    .pill {
      font-size: 12px;
      padding: 0 12px 0 8px;
    }

    .tagline-brand {
      font-size: 12px;
    }

    .install-pill {
      padding-left: 12px;
    }

    .copy-btn {
      height: 20px;
      padding: 0 6px;
    }

    .author-label {
      max-width: none;
    }

  }
</style>
