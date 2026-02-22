<script lang="ts">
  import { onMount } from 'svelte';

  const THEME = {
    bg:        '#1a1b26',
    fg:        '#c0caf5',
    selection: '#33467c',
  };

  let containerEl: HTMLDivElement;
  let reservedHeightPx = 512;

  function clamp(value: number, min: number, max: number): number {
    return Math.max(min, Math.min(value, max));
  }

  function getInitialTerminalSize(container: HTMLDivElement) {
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.visualViewport?.height ?? window.innerHeight;
    const isMobile = viewportWidth <= 640;
    const measuredWidth = Math.floor(container.getBoundingClientRect().width);
    const horizontalPadding = isMobile ? 6 : 4;
    const availableWidth = Math.max(280, (measuredWidth || viewportWidth) - horizontalPadding);
    const fontSize = isMobile ? 12 : 14;
    const cellWidth = isMobile ? 7.05 : 8.15;
    const cols = clamp(Math.floor((availableWidth - 4) / cellWidth), 32, 110);

    const chromeHeight = isMobile ? 260 : 230;
    const cellHeight = isMobile ? 14 : 16;
    const rows = clamp(Math.floor((viewportHeight - chromeHeight) / cellHeight), 16, 28);

    return { cols, rows, fontSize };
  }

  onMount(() => {
    let term: import('@xterm/xterm').Terminal | null = null;
    let observer: IntersectionObserver | null = null;
    let disposed = false;
    let removeMouseHandlers: (() => void) | null = null;

    const start = async () => {
      const initialSize = getInitialTerminalSize(containerEl);

      const { Terminal } = await import('@xterm/xterm');
      await import('@xterm/xterm/css/xterm.css');
      if (disposed) return;

      // wasm_exec.js is loaded via <script> in app.html
      const go = new (globalThis as any).Go();
      const result = await WebAssembly.instantiateStreaming(
        fetch('/bontree.wasm'),
        go.importObject
      );
      if (disposed) return;
      go.run(result.instance);

      const { cols, rows, fontSize } = initialSize;

      term = new Terminal({
        rows,
        cols,
        fontFamily: "'Maple Mono NF', 'MapleMono-NF-Regular', 'Symbols Nerd Font Mono', 'FiraCode Nerd Font', 'JetBrainsMono Nerd Font', monospace",
        fontSize,
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
      function mapKey(key: string, _domEvent: KeyboardEvent): string | null {
        // Arrow keys come as escape sequences from xterm
        if (key === '\x1b[A') return 'up';
        if (key === '\x1b[B') return 'down';
        if (key === '\x1b[C') return 'right';
        if (key === '\x1b[D') return 'left';
        if (key === '\r') return 'enter';
        if (key === '\x1b') return 'esc';
        if (key === '\x7f') return 'backspace';
        if (key === '\x04') return 'ctrl+d';
        if (key === '\x15') return 'ctrl+u';
        if (key === '\x06') return 'ctrl+f';
        // Regular printable characters
        if (key.length === 1 && key >= ' ') return key;
        return null;
      }

      function writeView(view: string) {
        if (!term) return;
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

      const onMouseDown = (e: MouseEvent) => {
        if (!term) return;
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
      };

      const onWheel = (e: WheelEvent) => {
        e.preventDefault();
        const dir = e.deltaY > 0 ? 1 : -1;
        const view = g.bontreeScroll(dir);
        writeView(view);
      };

      containerEl.addEventListener('mousedown', onMouseDown);
      containerEl.addEventListener('wheel', onWheel, { passive: false });
      removeMouseHandlers = () => {
        containerEl.removeEventListener('mousedown', onMouseDown);
        containerEl.removeEventListener('wheel', onWheel);
      };

      // Auto-focus
      observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting) setTimeout(() => term?.focus(), 300);
        },
        { threshold: 0.5 }
      );
      observer.observe(containerEl);
    };

    void start();

    return () => {
      disposed = true;
      removeMouseHandlers?.();
      observer?.disconnect();
      term?.dispose();
    };
  });
</script>

<div class="terminal-wrapper" style={`min-height: ${reservedHeightPx}px`}>
  <div class="terminal" bind:this={containerEl}></div>
</div>

<style>
  .terminal-wrapper {
    width: 100%;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid #414868;
  }

  .terminal {
    width: 100%;
    background: #1a1b26;
    padding: 4px 0;
    user-select: none;
    -webkit-user-select: none;
  }

  .terminal :global(.xterm),
  .terminal :global(.xterm-viewport),
  .terminal :global(.xterm-screen) {
    max-width: 100%;
  }
</style>
