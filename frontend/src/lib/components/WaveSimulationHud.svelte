<script lang="ts">
	import { sismoState } from '$lib/sismoStore.svelte';
	import {
		parseDepthKm,
		calculateEpicentralRadius,
		calculateDamageRadius,
		calculateLeadTime,
		formatTimeMMSS
	} from '$lib/utils/seismicWaves';
	import { Play, Pause, RotateCcw, X, Activity, ShieldAlert, FastForward, Navigation } from '@lucide/svelte';

	interface Props {
		onCenterEpicenter?: () => void;
	}

	let { onCenterEpicenter }: Props = $props();

	let lastTimestamp = 0;
	let animFrameId: number | null = null;

	const sim = $derived(sismoState.waveSimulation);
	const eq = $derived(sim.earthquake);

	const depth = $derived(eq ? parseDepthKm(eq.depth) : 10);
	const mag = $derived(eq?.magnitude || 5.0);
	const rPKm = $derived(calculateEpicentralRadius(sim.pSpeed, sim.timeSec, depth));
	const rSKm = $derived(calculateEpicentralRadius(sim.sSpeed, sim.timeSec, depth));
	const damageRadiusKm = $derived(calculateDamageRadius(mag, depth));
	const sampleLeadTime = $derived(calculateLeadTime(100, sim.pSpeed, sim.sSpeed));

	// Motor de simulación en tiempo real con requestAnimationFrame
	$effect(() => {
		if (sim.active && sim.isPlaying) {
			lastTimestamp = performance.now();
			const loop = (now: number) => {
				const deltaSec = (now - lastTimestamp) / 1000;
				lastTimestamp = now;

				if (sismoState.waveSimulation.isPlaying) {
					const nextTime = sismoState.waveSimulation.timeSec + deltaSec * sismoState.waveSimulation.speedMultiplier;
					if (nextTime >= sismoState.waveSimulation.maxTimeSec) {
						sismoState.waveSimulation.timeSec = sismoState.waveSimulation.maxTimeSec;
						sismoState.waveSimulation.isPlaying = false;
					} else {
						sismoState.waveSimulation.timeSec = nextTime;
						animFrameId = requestAnimationFrame(loop);
					}
				}
			};
			animFrameId = requestAnimationFrame(loop);
		} else {
			if (animFrameId) {
				cancelAnimationFrame(animFrameId);
				animFrameId = null;
			}
		}

		return () => {
			if (animFrameId) {
				cancelAnimationFrame(animFrameId);
				animFrameId = null;
			}
		};
	});

	function togglePlay() {
		if (sim.timeSec >= sim.maxTimeSec) {
			sim.timeSec = 0;
		}
		sim.isPlaying = !sim.isPlaying;
	}

	function resetSim() {
		sim.timeSec = 0;
		sim.isPlaying = true;
	}

	function setSpeed(mult: number) {
		sim.speedMultiplier = mult;
	}

	function closeSim() {
		sim.active = false;
		sim.isPlaying = false;
		sim.timeSec = 0;
	}

	function handleScrubber(e: Event) {
		const target = e.target as HTMLInputElement;
		sim.timeSec = parseFloat(target.value);
	}
</script>

{#if sim.active && eq}
	<div class="wave-hud-container glass">
		<!-- CABECERA DEL HUD -->
		<div class="hud-topbar">
			<div class="hud-badge">
				<span class="hud-dot"></span>
				<span class="hud-title">SIMULADOR GEOFÍSICO • ONDAS SÍSMICAS (P y S)</span>
			</div>
			<div class="hud-top-actions">
				{#if onCenterEpicenter}
					<button class="hud-mini-btn" onclick={onCenterEpicenter} title="Centrar Epicentro">
						<Navigation size={13} />
						<span>Centrar</span>
					</button>
				{/if}
				<button class="hud-close-btn" onclick={closeSim} title="Cerrar Simulador">
					<X size={16} />
				</button>
			</div>
		</div>

		<!-- INFORMACIÓN DEL EVENTO -->
		<div class="hud-event-row">
			<div class="event-meta">
				<span class="event-mag" style="color: {mag >= 6 ? '#ef4444' : mag >= 4.5 ? '#f59e0b' : '#10b981'}">
					M{mag.toFixed(1)}
				</span>
				<span class="event-loc">{eq.location}</span>
			</div>
			<div class="event-depth">Profundidad: <strong>{depth} km</strong></div>
		</div>

		<!-- PANEL DE TELEMETRÍA -->
		<div class="hud-telemetry-grid">
			<!-- Cronómetro T+ -->
			<div class="telemetry-box chrono-box">
				<div class="telemetry-label">TIEMPO EPICENTRAL</div>
				<div class="chrono-value">T + {formatTimeMMSS(sim.timeSec)}</div>
				<div class="telemetry-sub">Ventana de propagación</div>
			</div>

			<!-- Onda P -->
			<div class="telemetry-box p-box">
				<div class="telemetry-label">
					<span class="wave-tag p-tag">ONDA P</span> PRIMARIA
				</div>
				<div class="telemetry-val p-val">
					{rPKm > 0 ? `${rPKm.toFixed(0)} km` : 'En corteza...'}
				</div>
				<div class="telemetry-sub">v = {sim.pSpeed} km/s • Compresional</div>
			</div>

			<!-- Onda S -->
			<div class="telemetry-box s-box">
				<div class="telemetry-label">
					<span class="wave-tag s-tag">ONDA S</span> DESTRUCTIVA
				</div>
				<div class="telemetry-val s-val">
					{rSKm > 0 ? `${rSKm.toFixed(0)} km` : 'En corteza...'}
				</div>
				<div class="telemetry-sub">v = {sim.sSpeed} km/s • Cizalla</div>
			</div>

			<!-- Alerta Temprana (EEW) -->
			<div class="telemetry-box eew-box">
				<div class="telemetry-label">VENTAJA EEW A 100 KM</div>
				<div class="telemetry-val eew-val">+{sampleLeadTime}s</div>
				<div class="telemetry-sub">
					{damageRadiusKm > 0 ? `Radio Daño: ${damageRadiusKm} km` : 'Daño Estructural Leve'}
				</div>
			</div>
		</div>

		<!-- SCRUBBER / BARRA DE TIEMPO INTERACTIVA -->
		<div class="hud-scrubber-wrapper">
			<span class="scrubber-time">0s</span>
			<input
				type="range"
				min="0"
				max={sim.maxTimeSec}
				step="0.5"
				value={sim.timeSec}
				oninput={handleScrubber}
				class="hud-range"
			/>
			<span class="scrubber-time">{sim.maxTimeSec}s</span>
		</div>

		<!-- BARRA DE CONTROLES INFERIOR -->
		<div class="hud-controls-bar">
			<div class="controls-playback">
				<button class="hud-action-btn primary-btn" onclick={togglePlay}>
					{#if sim.isPlaying}
						<Pause size={15} />
						<span>Pausar</span>
					{:else}
						<Play size={15} />
						<span>{sim.timeSec >= sim.maxTimeSec ? 'Repetir' : 'Reproducir'}</span>
					{/if}
				</button>
				<button class="hud-action-btn" onclick={resetSim} title="Reiniciar a 0s">
					<RotateCcw size={15} />
					<span>Reiniciar</span>
				</button>
			</div>

			<div class="controls-speed">
				<span class="speed-label"><FastForward size={13} /> Velocidad:</span>
				{#each [1, 2, 5] as mult}
					<button
						class="speed-btn {sim.speedMultiplier === mult ? 'active' : ''}"
						onclick={() => setSpeed(mult)}
					>
						{mult}x
					</button>
				{/each}
			</div>

			<div class="hud-legend-hint">
				<span class="legend-chip chip-p">● Onda P (Cyan)</span>
				<span class="legend-chip chip-s">● Onda S (Rojo)</span>
			</div>
		</div>
	</div>
{/if}

<style>
	.wave-hud-container {
		position: absolute;
		bottom: 24px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 1100;
		width: min(92%, 760px);
		background: rgba(10, 15, 29, 0.94);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
		border: 1px solid rgba(0, 240, 255, 0.35);
		box-shadow: 0 16px 36px rgba(0, 0, 0, 0.65), 0 0 20px rgba(0, 240, 255, 0.15);
		border-radius: 12px;
		padding: 14px 18px;
		color: #f8fafc;
		font-family: inherit;
		animation: hud-appear 0.3s cubic-bezier(0.16, 1, 0.3, 1);
	}

	@keyframes hud-appear {
		from {
			opacity: 0;
			transform: translate(-50%, 20px);
		}
		to {
			opacity: 1;
			transform: translate(-50%, 0);
		}
	}

	.hud-topbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-bottom: 8px;
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}

	.hud-badge {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.hud-dot {
		width: 8px;
		height: 8px;
		background: #00f0ff;
		border-radius: 50%;
		box-shadow: 0 0 10px #00f0ff;
		animation: pulse-dot 1.5s infinite;
	}

	@keyframes pulse-dot {
		0%, 100% { opacity: 1; transform: scale(1); }
		50% { opacity: 0.4; transform: scale(0.85); }
	}

	.hud-title {
		font-size: 11px;
		font-weight: 800;
		letter-spacing: 0.8px;
		color: #00f0ff;
		text-transform: uppercase;
	}

	.hud-top-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.hud-mini-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		background: rgba(255, 255, 255, 0.06);
		border: 1px solid rgba(255, 255, 255, 0.15);
		color: #e2e8f0;
		padding: 3px 8px;
		border-radius: 4px;
		font-size: 11px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.hud-mini-btn:hover {
		background: rgba(0, 240, 255, 0.15);
		border-color: #00f0ff;
		color: #00f0ff;
	}

	.hud-close-btn {
		background: transparent;
		border: none;
		color: #94a3b8;
		cursor: pointer;
		padding: 4px;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: color 0.2s;
	}

	.hud-close-btn:hover {
		color: #ef4444;
	}

	.hud-event-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin: 8px 0 12px 0;
		font-size: 13px;
	}

	.event-meta {
		display: flex;
		align-items: center;
		gap: 8px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.event-mag {
		font-weight: 800;
		font-size: 16px;
	}

	.event-loc {
		color: #e2e8f0;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.event-depth {
		color: #94a3b8;
		font-size: 12px;
		white-space: nowrap;
	}

	.hud-telemetry-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 10px;
		margin-bottom: 12px;
	}

	@media (max-width: 640px) {
		.hud-telemetry-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	.telemetry-box {
		background: rgba(15, 23, 42, 0.65);
		border: 1px solid rgba(255, 255, 255, 0.08);
		border-radius: 8px;
		padding: 8px 10px;
		text-align: center;
		transition: border-color 0.2s;
	}

	.telemetry-label {
		font-size: 10px;
		font-weight: 700;
		color: #94a3b8;
		letter-spacing: 0.5px;
		margin-bottom: 2px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
	}

	.chrono-value {
		font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
		font-size: 20px;
		font-weight: 800;
		color: #38bdf8;
		text-shadow: 0 0 10px rgba(56, 189, 248, 0.3);
	}

	.telemetry-val {
		font-size: 18px;
		font-weight: 800;
	}

	.p-val {
		color: #00f0ff;
		text-shadow: 0 0 10px rgba(0, 240, 255, 0.35);
	}

	.s-val {
		color: #ef4444;
		text-shadow: 0 0 10px rgba(239, 68, 68, 0.35);
	}

	.eew-val {
		color: #10b981;
		text-shadow: 0 0 10px rgba(16, 185, 129, 0.35);
	}

	.telemetry-sub {
		font-size: 9.5px;
		color: #64748b;
		margin-top: 2px;
	}

	.wave-tag {
		font-size: 9px;
		padding: 1px 4px;
		border-radius: 3px;
		font-weight: 800;
	}

	.p-tag {
		background: rgba(0, 240, 255, 0.15);
		color: #00f0ff;
	}

	.s-tag {
		background: rgba(239, 68, 68, 0.15);
		color: #ef4444;
	}

	.hud-scrubber-wrapper {
		display: flex;
		align-items: center;
		gap: 10px;
		margin-bottom: 12px;
	}

	.scrubber-time {
		font-size: 11px;
		font-family: monospace;
		color: #94a3b8;
		min-width: 28px;
	}

	.hud-range {
		flex: 1;
		accent-color: #00f0ff;
		cursor: pointer;
		height: 6px;
	}

	.hud-controls-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 10px;
		padding-top: 6px;
		border-top: 1px solid rgba(255, 255, 255, 0.08);
	}

	.controls-playback {
		display: flex;
		gap: 8px;
	}

	.hud-action-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 14px;
		border-radius: 6px;
		font-size: 12px;
		font-weight: 700;
		cursor: pointer;
		border: 1px solid rgba(255, 255, 255, 0.15);
		background: rgba(255, 255, 255, 0.06);
		color: #f1f5f9;
		transition: all 0.2s;
	}

	.hud-action-btn:hover {
		background: rgba(255, 255, 255, 0.12);
	}

	.primary-btn {
		background: #00f0ff;
		color: #0b1120;
		border-color: #00f0ff;
		box-shadow: 0 0 12px rgba(0, 240, 255, 0.4);
	}

	.primary-btn:hover {
		background: #38bdf8;
		box-shadow: 0 0 16px rgba(0, 240, 255, 0.6);
	}

	.controls-speed {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
	}

	.speed-label {
		display: flex;
		align-items: center;
		gap: 4px;
		color: #94a3b8;
		margin-right: 4px;
	}

	.speed-btn {
		background: rgba(255, 255, 255, 0.06);
		border: 1px solid rgba(255, 255, 255, 0.12);
		color: #94a3b8;
		padding: 2px 7px;
		border-radius: 4px;
		font-size: 11px;
		font-weight: 700;
		cursor: pointer;
		transition: all 0.2s;
	}

	.speed-btn.active {
		background: rgba(0, 240, 255, 0.2);
		border-color: #00f0ff;
		color: #00f0ff;
	}

	.hud-legend-hint {
		display: flex;
		gap: 8px;
		font-size: 11px;
	}

	.legend-chip {
		font-weight: 600;
	}

	.chip-p { color: #00f0ff; }
	.chip-s { color: #ef4444; }
</style>
