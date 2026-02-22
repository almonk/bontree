<script lang="ts">
  import { onMount } from 'svelte';

  const THEME = {
    bg:        '#1a1b26',
    fg:        '#c0caf5',
    selection: '#33467c',
  };

  let containerEl: HTMLDivElement;

  onMount(async () => {
    const { Terminal } = await import('@xterm/xterm');
    await import('@xterm/xterm/css/xterm.css');

    // wasm_exec.js is loaded via <script> in app.html
    const go = new (globalThis as any).Go();
    const result = await WebAssembly.instantiateStreaming(
      fetch('/bontree.wasm'),
      go.importObject
    );
    go.run(result.instance);

    const term = new Terminal({
      rows: 24,
      cols: 80,
      fontFamily: "'Maple Mono NF', 'MapleMono-NF-Regular', 'Symbols Nerd Font Mono', 'FiraCode Nerd Font', 'JetBrainsMono Nerd Font', monospace",
      fontSize: 14,
      lineHeight: 1.0,
      theme: {
        background: THEME.bg,
        foreground: THEME.fg,
        cursor: THEME.bg,
        cursorAccent: THEME.bg,
        selectionBackground: THEME.selection,
      },
      fontWeightBold: 'normal',
      cursorBlink: false,
      cursorStyle: 'bar',
      cursorInactiveStyle: 'none',
      allowTransparency: false,
      scrollback: 0,
      disableStdin: false,
    });

    term.open(containerEl);

    // Get global bridge functions from WASM
    const g = globalThis as any;

    // Map xterm key sequences to bontree key names
    function mapKey(key: string, domEvent: KeyboardEvent): string | null {
      // Arrow keys come as escape sequences from xterm
      if (key === '\x1b[A') return 'up';
      if (key === '\x1b[B') return 'down';
      if (key === '\x1b[C') return 'right';
      if (key === '\x1b[D') return 'left';
      if (key === '\r')     return 'enter';
      if (key === '\x1b')   return 'esc';
      if (key === '\x7f')   return 'backspace';
      if (key === '\x04')   return 'ctrl+d';
      if (key === '\x15')   return 'ctrl+u';
      if (key === '\x06')   return 'ctrl+f';
      // Regular printable characters
      if (key.length === 1 && key >= ' ') return key;
      return null;
    }

    function writeView(view: string) {
      term.write('\x1b[?25l\x1b[H');
      // Split into lines and write with clear-to-end
      const lines = view.split('\n');
      for (let i = 0; i < term.rows; i++) {
        term.write('\x1b[K');
        if (i < lines.length) term.write(lines[i]);
        if (i < term.rows - 1) term.write('\r\n');
      }
    }

    // Initial render
    const initial = g.bontreeInit(term.cols, term.rows);
    writeView(initial);

    // Keyboard input
    term.onKey(({ key, domEvent }: { key: string; domEvent: KeyboardEvent }) => {
      const mapped = mapKey(key, domEvent);
      if (!mapped) return;

      const result = g.bontreeKey(mapped);
      if (typeof result === 'string') {
        writeView(result);
      } else if (result && result.view) {
        writeView(result.view);
        if (result.flash) {
          setTimeout(() => {
            const cleared = g.bontreeClearFlash();
            writeView(cleared);
          }, 2000);
        }
      }
    });

    // Mouse: click / double-click
    let lastClickTime = 0;
    let lastClickRow = -1;

    containerEl.addEventListener('mousedown', (e) => {
      term.focus();
      const termEl = containerEl.querySelector('.xterm-rows');
      if (!termEl) return;
      const rect = termEl.getBoundingClientRect();
      const cellHeight = rect.height / term.rows;
      const row = Math.floor((e.clientY - rect.top) / cellHeight);
      const vh = term.rows - 1; // approximate viewport
      if (row < 0 || row >= vh) return;

      const now = Date.now();
      const isDouble = now - lastClickTime < 400 && row === lastClickRow;

      const view = g.bontreeClick(row, isDouble);
      writeView(view);

      if (isDouble) {
        lastClickTime = 0;
        lastClickRow = -1;
      } else {
        lastClickTime = now;
        lastClickRow = row;
      }
    });

    // Mouse: scroll
    containerEl.addEventListener('wheel', (e) => {
      e.preventDefault();
      const dir = e.deltaY > 0 ? 1 : -1;
      const view = g.bontreeScroll(dir);
      writeView(view);
    }, { passive: false });

    // Auto-focus
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) setTimeout(() => term.focus(), 300);
    }, { threshold: 0.5 });
    observer.observe(containerEl);

    return () => {
      observer.disconnect();
      term.dispose();
    };
  });
</script>

<div class="terminal-wrapper">
  <div class="terminal-chrome">
    <div class="terminal-dots">
      <span class="dot red"></span>
      <span class="dot yellow"></span>
      <span class="dot green"></span>
    </div>
    <span class="terminal-title">bontree ~/my-project</span>
  </div>
  <div class="terminal" bind:this={containerEl}></div>
</div>

<style>
  .terminal-wrapper {
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--border);
  }

  .terminal-chrome {
    background: #16161e;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .terminal-dots {
    display: flex;
    gap: 8px;
  }

  .dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
  }
  .dot.red    { background: #f7768e; }
  .dot.yellow { background: #e0af68; }
  .dot.green  { background: #9ece6a; }

  .terminal-title {
    color: var(--fg-muted);
    font-size: 0.8rem;
    flex: 1;
    text-align: center;
    margin-right: 52px;
  }

  .terminal {
    background: #1a1b26;
    padding: 4px 0;
    user-select: none;
    -webkit-user-select: none;
  }
</style>
