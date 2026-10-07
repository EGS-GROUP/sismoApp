<script lang="ts">
	import type { Earthquake } from '$lib/types';
	import { getM3U8ProxyUrl } from '$lib/api';

	interface Props {
		selectedEarthquake: Earthquake | null;
		activeTab: string;
	}

	let { selectedEarthquake, activeTab }: Props = $props();

	let tvVideoRef: HTMLVideoElement | null = $state(null);
	let tvIframeRef: HTMLIFrameElement | null = $state(null);
	let hlsInstance: any = null;
	let tvFooterMessage = $state('Las cadenas oficiales pueden bloquear la reproducción externa por derechos de autor.');
	let tvSourceLink = $state('');
	let showTvExternalLink = $state(false);

	const FALLBACK_DM_ID = 'x74272f';
	const DEFAULT_HLS_URL = 'https://rt-esp.rttv.com/live/rtesp/playlist.m3u8';
	const DEFAULT_SOURCE_LINK = 'https://actualidad.rt.com/en_vivo';

	let tvChannelId = $derived.by(() => {
		if (!selectedEarthquake) return '';
		const loc = selectedEarthquake.location.toLowerCase();
		if (loc.includes('venezuela')) return 'DM:x930kre';
		if (loc.includes('chile')) return 'M3U8:https://unlimited1-cl-isp.dps.live/24horas/24horas.smil/playlist.m3u8';
		if (loc.includes('colombia')) return 'M3U8:https://streaming.rtvc.gov.co/TV_Canal_Institucional/chunklist.m3u8';
		if (loc.includes('peru')) return 'M3U8:https://cdns.rpp.pe/tv/rpptv/playlist.m3u8';
		if (loc.includes('mexico')) return 'DM:x8k1m3t';
		return `M3U8:${DEFAULT_HLS_URL}`;
	});

	let tvSourceUrl = $derived.by(() => {
		if (!selectedEarthquake) return '';
		const loc = selectedEarthquake.location.toLowerCase();
		if (loc.includes('venezuela')) return 'https://vtv.gob.ve/';
		if (loc.includes('chile')) return 'https://www.24horas.cl/';
		if (loc.includes('colombia')) return 'https://www.canalinstitucional.tv/';
		if (loc.includes('peru')) return 'https://rpp.pe/tv';
		if (loc.includes('mexico')) return 'https://www.milenio.com/';
		return DEFAULT_SOURCE_LINK;
	});

	function stopAllTV() {
		if (hlsInstance) { hlsInstance.destroy(); hlsInstance = null; }
		if (tvVideoRef) {
			tvVideoRef.pause();
			tvVideoRef.src = '';
			tvVideoRef.load();
			tvVideoRef.style.display = 'none';
		}
		if (tvIframeRef) {
			tvIframeRef.src = '';
			tvIframeRef.style.display = 'none';
		}
	}

	function triggerDegradation() {
		console.warn('MIA-C4I: Stream failed, degrading to fallback');
		stopAllTV();
		if (tvIframeRef) {
			tvIframeRef.style.display = 'block';
			tvIframeRef.src = `https://www.dailymotion.com/embed/video/${FALLBACK_DM_ID}?autoplay=1&mute=1`;
		}
		tvFooterMessage = 'Conexión principal falló. Mostrando transmisión internacional de respaldo.';
		showTvExternalLink = true;
	}

	function initLiveTV() {
		if (!tvVideoRef) return;
		const Hls = (window as any).Hls;
		if (Hls && Hls.isSupported()) {
			hlsInstance = new Hls();
			hlsInstance.loadSource(getM3U8ProxyUrl(DEFAULT_HLS_URL));
			hlsInstance.attachMedia(tvVideoRef);
			hlsInstance.on(Hls.Events.MANIFEST_PARSED, function() {
				if (tvVideoRef) tvVideoRef.play().catch(() => {});
			});
		} else if (tvVideoRef.canPlayType('application/vnd.apple.mpegurl')) {
			tvVideoRef.src = getM3U8ProxyUrl(DEFAULT_HLS_URL);
			tvVideoRef.play().catch(() => {});
		}
	}

	$effect(() => {
		if (activeTab === 'tv' && tvChannelId) {
			tvFooterMessage = 'Estableciendo conexión con la señal en vivo...';
			showTvExternalLink = false;
			tvSourceLink = tvSourceUrl;
			stopAllTV();

			if (tvChannelId.startsWith('M3U8:') && tvVideoRef) {
				tvVideoRef.style.display = 'block';
				const rawUrl = tvChannelId.substring(5);
				const proxyUrl = getM3U8ProxyUrl(rawUrl);
				const Hls = (window as any).Hls;
				if (Hls && Hls.isSupported()) {
					hlsInstance = new Hls({ manifestLoadingTimeOut: 20000 });
					hlsInstance.loadSource(proxyUrl);
					hlsInstance.attachMedia(tvVideoRef);
					hlsInstance.on(Hls.Events.MANIFEST_PARSED, () => {
						tvFooterMessage = 'Conexión HLS establecida. Señal abierta.';
						tvVideoRef?.play().catch(e => console.log('Autoplay prevented', e));
					});
					hlsInstance.on(Hls.Events.ERROR, (_e: any, data: any) => {
						if (data.fatal) triggerDegradation();
					});
				} else if (tvVideoRef.canPlayType('application/vnd.apple.mpegurl')) {
					tvVideoRef.src = proxyUrl;
					tvVideoRef.play().catch(e => console.log('Autoplay prevented', e));
				}
			} else if (tvChannelId.startsWith('DM:') && tvIframeRef) {
				const dmId = tvChannelId.substring(3);
				tvIframeRef.src = `https://www.dailymotion.com/embed/video/${dmId}?autoplay=1`;
				tvIframeRef.style.display = 'block';
				tvFooterMessage = 'Conexión Dailymotion establecida.';
			}
		} else if (activeTab === 'tv' && !tvChannelId && tvVideoRef && !hlsInstance) {
			initLiveTV();
		} else {
			stopAllTV();
		}
	});
</script>

<div class="tab-content active live-tv-container">
	<div class="live-tv-video">
		<video bind:this={tvVideoRef} controls style="width: 100%; height: 100%; object-fit: contain; display: none;"></video>
		<iframe bind:this={tvIframeRef} title="Live TV" style="width: 100%; height: 100%; border: none; display: none;" allow="autoplay; encrypted-media" allowfullscreen></iframe>
	</div>
	<div class="live-tv-info">
		<div style="text-align: center;">
			<div style="font-size: 11px; color: var(--text-secondary); text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; margin-bottom: 4px;">Señal</div>
			<div style="font-size: 13px; color: var(--text-primary); font-weight: 700;">{tvChannelId ? 'Regional' : 'Internacional'}</div>
		</div>
		<div style="text-align: center; font-size: 11px; color: var(--text-secondary); line-height: 1.4;">
			{tvFooterMessage}
		</div>
		{#if showTvExternalLink}
			<a href={tvSourceLink} target="_blank" style="text-align: center; color: var(--accent); font-size: 12px; text-decoration: none; border: 1px solid var(--accent); padding: 8px 10px; border-radius: 4px; display: inline-flex; align-items: center; justify-content: center; gap: 4px; background: color-mix(in srgb, var(--accent) 10%, transparent);">Ver Fuente</a>
		{/if}
	</div>
</div>
