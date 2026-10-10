/**
 * Motor de Sonificación Sísmica (Web Audio API)
 * Transpone las frecuencias infrasónicas del sismo al espectro audible humano en tiempo real.
 * Cero dependencias externas.
 */

let audioCtx: AudioContext | null = null;
let masterGain: GainNode | null = null;

// Nodos para Onda S (Rumble / Sacudida subterránea)
let sOsc1: OscillatorNode | null = null;
let sOsc2: OscillatorNode | null = null;
let sGain: GainNode | null = null;
let sFilter: BiquadFilterNode | null = null;
let noiseNode: AudioBufferSourceNode | null = null;
let noiseGain: GainNode | null = null;

let isAudioRunning = false;
let audioEnabled = false;

function createNoiseBuffer(ctx: AudioContext): AudioBuffer {
	const bufferSize = ctx.sampleRate * 2; // 2 segundos de buffer en loop
	const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
	const data = buffer.getChannelData(0);
	let lastOut = 0.0;

	// Generador de ruido marrón/rosa (simula fricción tectónica)
	for (let i = 0; i < bufferSize; i++) {
		const white = Math.random() * 2 - 1;
		lastOut = (lastOut + 0.02 * white) / 1.02;
		data[i] = lastOut * 3.5;
	}
	return buffer;
}

export function initAudio(): boolean {
	if (typeof window === 'undefined') return false;

	try {
		const AudioContextClass = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
		if (!AudioContextClass) return false;

		if (!audioCtx) {
			audioCtx = new AudioContextClass();
		}

		if (audioCtx.state === 'suspended') {
			audioCtx.resume();
		}

		if (!masterGain) {
			masterGain = audioCtx.createGain();
			masterGain.gain.setValueAtTime(0.3, audioCtx.currentTime);
			masterGain.connect(audioCtx.destination);
		}

		return true;
	} catch (e) {
		console.warn('[SeismicAudio] No se pudo inicializar AudioContext:', e);
		return false;
	}
}

export function startSeismicRumble(mag = 5.0) {
	if (!audioCtx || !masterGain || isAudioRunning) return;

	try {
		const now = audioCtx.currentTime;

		// 1. Filtro paso-bajo para el rumble sordo
		sFilter = audioCtx.createBiquadFilter();
		sFilter.type = 'lowpass';
		sFilter.frequency.setValueAtTime(90, now);
		sFilter.Q.setValueAtTime(3.0, now);

		// 2. Ganancia de Onda S
		sGain = audioCtx.createGain();
		sGain.gain.setValueAtTime(0.001, now);
		sGain.gain.linearRampToValueAtTime(Math.min(0.5, 0.15 + (mag - 4) * 0.08), now + 0.3);

		// 3. Osciladores duales de baja frecuencia (subarmónicos 40-58 Hz)
		sOsc1 = audioCtx.createOscillator();
		sOsc1.type = 'sine';
		sOsc1.frequency.setValueAtTime(44, now);

		sOsc2 = audioCtx.createOscillator();
		sOsc2.type = 'triangle';
		sOsc2.frequency.setValueAtTime(56, now);

		// 4. Ruido tectónico filtrado
		const noiseBuf = createNoiseBuffer(audioCtx);
		noiseNode = audioCtx.createBufferSource();
		noiseNode.buffer = noiseBuf;
		noiseNode.loop = true;

		noiseGain = audioCtx.createGain();
		noiseGain.gain.setValueAtTime(0.08, now);

		// Conectar cadenas
		sOsc1.connect(sFilter);
		sOsc2.connect(sFilter);
		sFilter.connect(sGain);

		noiseNode.connect(sFilter);

		sGain.connect(masterGain);

		// Iniciar osciladores
		sOsc1.start(now);
		sOsc2.start(now);
		noiseNode.start(now);

		isAudioRunning = true;
	} catch (e) {
		console.warn('[SeismicAudio] Error iniciando rumble:', e);
	}
}

/**
 * Emite un pulso acústico inicial al propagarse la onda P (compresional)
 */
export function playPWaveChirp() {
	if (!audioCtx || !masterGain || !audioEnabled) return;

	try {
		const now = audioCtx.currentTime;
		const osc = audioCtx.createOscillator();
		const gain = audioCtx.createGain();

		osc.type = 'sine';
		// Barrido de frecuencia descendente (260 Hz -> 140 Hz)
		osc.frequency.setValueAtTime(260, now);
		osc.frequency.exponentialRampToValueAtTime(140, now + 0.25);

		gain.gain.setValueAtTime(0.2, now);
		gain.gain.exponentialRampToValueAtTime(0.001, now + 0.25);

		osc.connect(gain);
		gain.connect(masterGain);

		osc.start(now);
		osc.stop(now + 0.26);
	} catch (_) {}
}

/**
 * Modula dinámicamente la ganancia y frecuencia del audio con base en la simulación
 */
export function updateAudioFrame(timeSec: number, rPKm: number, rSKm: number, isPlaying: boolean, mag = 5.0) {
	if (!audioEnabled) {
		if (isAudioRunning) stopAudio();
		return;
	}

	if (!isPlaying || timeSec <= 0) {
		if (isAudioRunning) stopAudio();
		return;
	}

	if (!isAudioRunning && rSKm > 0) {
		startSeismicRumble(mag);
	}

	if (isAudioRunning && sGain && sFilter && audioCtx) {
		const now = audioCtx.currentTime;
		// Modulación suave: a mayor tiempo y distancia, atenuación física
		const attenuation = Math.max(0.05, 1 / (1 + rSKm * 0.005));
		const targetVol = Math.min(0.4, 0.25 * attenuation);
		sGain.gain.setTargetAtTime(targetVol, now, 0.1);

		// Pequeña oscilación sismológica en el filtro
		const wobble = 85 + Math.sin(timeSec * 4) * 15;
		sFilter.frequency.setTargetAtTime(wobble, now, 0.1);
	}
}

export function stopAudio() {
	if (!isAudioRunning || !audioCtx) return;

	try {
		const now = audioCtx.currentTime;
		if (sGain) {
			sGain.gain.linearRampToValueAtTime(0.0001, now + 0.15);
		}

		setTimeout(() => {
			if (sOsc1) {
				try { sOsc1.stop(); sOsc1.disconnect(); } catch (_) {}
				sOsc1 = null;
			}
			if (sOsc2) {
				try { sOsc2.stop(); sOsc2.disconnect(); } catch (_) {}
				sOsc2 = null;
			}
			if (noiseNode) {
				try { noiseNode.stop(); noiseNode.disconnect(); } catch (_) {}
				noiseNode = null;
			}
			isAudioRunning = false;
		}, 160);
	} catch (_) {
		isAudioRunning = false;
	}
}

export function setAudioEnabled(enabled: boolean): boolean {
	audioEnabled = enabled;
	if (enabled) {
		initAudio();
		playPWaveChirp();
	} else {
		stopAudio();
	}
	return audioEnabled;
}

export function getAudioEnabled(): boolean {
	return audioEnabled;
}
