<script lang="ts">
	export let size: 'small' | 'medium' | 'large' = 'medium';
	export let color: 'primary' | 'secondary' | 'white' = 'primary';
</script>

<div
	class="book-loader"
	class:small={size === 'small'}
	class:medium={size === 'medium'}
	class:large={size === 'large'}
	class:primary={color === 'primary'}
	class:secondary={color === 'secondary'}
	class:white={color === 'white'}
	role="status"
	aria-live="polite"
	aria-label="Loading"
>
	<!--
		Turning leaves are split: all fill shapes first, then all stroke-only shapes.
		That way a neighbor's gold fill can never paint over another page's navy outline.
		Each fill/edge pair runs the same book-turn keyframes and delay.
	-->
	<svg class="book-loader__svg" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
		<!-- Base: fill and edge split so fill can be softened without washing out the stroke. -->
		<path class="book-loader__base-fill" d="M20 25 L80 25 L80 75 L20 75 Z" />
		<g class="book-loader__turn-fills" aria-hidden="true">
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-fill" d="M 50 25 L 80 25 L 80 75 L 50 75" />
		</g>
		<g class="book-loader__turn-edges" aria-hidden="true">
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
			<path class="book-loader__turn book-loader__turn-edge" d="M 50 25 L 80 25 L 80 75 L 50 75" />
		</g>
		<path class="book-loader__base-edge" d="M20 25 L80 25 L80 75 L20 75 Z" />
	</svg>
</div>

<style>
	/* ─── colour / token layer ─────────────────────────────────────── */
	.book-loader {
		display: flex;
		align-items: center;
		justify-content: center;
		line-height: 0;
		--T: 6s;
		--book-fill: #cca73c;
		--book-stroke: #102e50;
		
	}

	.book-loader.primary {
		--book-fill: #cca73c;
		--book-stroke: #102e50;
	}

	.book-loader.secondary {
		--book-fill: #cca73c;
		--book-stroke: color-mix(in srgb, var(--secondary) 35%, #102e50);
	}

	.book-loader.white {
		--book-fill: #cca73c;
		--book-stroke: rgba(16, 46, 80, 0.85);
	}

	/* ─── size layer ───────────────────────────────────────────────── */
	.book-loader__svg {
		display: block;
		width: 200px;
		height: 200px;
	}

	.book-loader.small .book-loader__svg  { width: 44px; height: 44px; }
	.book-loader.medium .book-loader__svg { width: 100px; height: 100px; }
	.book-loader.large  .book-loader__svg { width: 150px; height: 150px; }

	@media (max-width: 768px) {
		.book-loader.large .book-loader__svg { width: 56px; height: 56px; }
	}

	@media (max-width: 480px) {
		.book-loader.large .book-loader__svg { width: 48px; height: 48px; }
	}

	/* ─── animation layer ──────────────────────────────────────────── */
	.book-loader__base-fill {
		fill: rgb(136, 107, 20);
		fill-opacity: var(--book-page-fill-opacity, 1);
		stroke: none;
	}

	.book-loader__base-edge {
		fill: none;
		stroke: var(--book-stroke);
		stroke-width: 3;
		stroke-linejoin: round;
	}

	.book-loader__turn {
		opacity: 0;
		animation: book-turn var(--T) linear infinite;
	}

	.book-loader__turn-fill {
		fill: var(--book-fill);
		fill-opacity: var(--book-page-fill-opacity, 1);
		stroke: none;
	}

	.book-loader__turn-edge {
		fill: none;
		stroke: var(--book-stroke);
		stroke-width: 3;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	/*
	 * Burst + pause: first 25% of each path's cycle is the turn; rest parked hidden.
	 * Stagger pairs (fill + edge) by T/12.
	 */
	.book-loader__turn-fills > .book-loader__turn:nth-child(1),
	.book-loader__turn-edges > .book-loader__turn:nth-child(1) {
		animation-delay: calc(var(--T) * 0 / 12);
	}
	.book-loader__turn-fills > .book-loader__turn:nth-child(2),
	.book-loader__turn-edges > .book-loader__turn:nth-child(2) {
		animation-delay: calc(var(--T) * 1 / 12);
	}
	.book-loader__turn-fills > .book-loader__turn:nth-child(3),
	.book-loader__turn-edges > .book-loader__turn:nth-child(3) {
		animation-delay: calc(var(--T) * 2 / 12);
	}
	.book-loader__turn-fills > .book-loader__turn:nth-child(4),
	.book-loader__turn-edges > .book-loader__turn:nth-child(4) {
		animation-delay: calc(var(--T) * 3 / 12);
	}
	.book-loader__turn-fills > .book-loader__turn:nth-child(5),
	.book-loader__turn-edges > .book-loader__turn:nth-child(5) {
		animation-delay: calc(var(--T) * 4 / 12);
	}
	.book-loader__turn-fills > .book-loader__turn:nth-child(6),
	.book-loader__turn-edges > .book-loader__turn:nth-child(6) {
		animation-delay: calc(var(--T) * 5 / 12);
	}

	/*
	 * @keyframes book-turn
	 *
	 * 0%–24%    : ACTIVE — page sweeps right→left with bow (real snapshot coords)
	 * 24%–25%   : fade out at flat-left
	 * 25%+      : instantly reset to flat-right (hidden), park until next cycle
	 * 100%      : flat-right at opacity 0, ready for next iteration
	 */
	@keyframes book-turn {
		/* ── begin turn at flat-right ── */
		0% {
			d: path("M 50 25 L 80 25 L 80 75 L 50 75");
			opacity: 1;
		}

		/* right side — approaching spine */
		2%    { d: path("M 50 25 L 76 24.4 L 76 75.6 L 50 75"); }
		4%    { d: path("M 50 25 L 71 23.5 L 71 76.5 L 50 75"); }
		6%    { d: path("M 50 25 L 66 22.6 L 66 77.4 L 50 75"); }
		8%    { d: path("M 50 25 L 61 21.9 L 61 78.1 L 50 75"); }
		10%   { d: path("M 50 25 L 56 21.0 L 56 79.0 L 50 75"); }

		/* peak bow at spine */
		12.5% { d: path("M 50 25 L 50 20.1 L 50 79.9 L 50 75"); }

		/* left side — leaving spine */
		14.5% { d: path("M 50 25 L 44 21.0 L 44 79.0 L 50 75"); }
		16.5% { d: path("M 50 25 L 39 21.8 L 39 78.2 L 50 75"); }
		18.5% { d: path("M 50 25 L 34 22.6 L 34 77.4 L 50 75"); }
		20.5% { d: path("M 50 25 L 29 23.4 L 29 76.6 L 50 75"); }
		22.5% { d: path("M 50 25 L 21 24.8 L 21 75.2 L 50 75"); }

		/* ── arrive flat-left, fade out ── */
		24% {
			d: path("M 50 25 L 20 25 L 20 75 L 50 75");
			opacity: 1;
		}
		25% {
			d: path("M 50 25 L 20 25 L 20 75 L 50 75");
			opacity: 0;
		}

		/* ── instantly reset to flat-right while hidden, park for the rest of the cycle ── */
		25.1% {
			d: path("M 50 25 L 80 25 L 80 75 L 50 75");
			opacity: 0;
		}
		100% {
			d: path("M 50 25 L 80 25 L 80 75 L 50 75");
			opacity: 0;
		}
	}

	/* ─── reduced-motion fallback ──────────────────────────────────── */
	@media (prefers-reduced-motion: reduce) {
		.book-loader__turn {
			animation: none;
			opacity: 0;
		}

		.book-loader__turn-fills > .book-loader__turn:nth-child(1),
		.book-loader__turn-edges > .book-loader__turn:nth-child(1) {
			d: path("M 50 25 L 50 20.1 L 50 79.9 L 50 75");
			opacity: 1;
		}
	}
</style>
