<script lang="ts">
	import { X, Globe, MapPin, Radio, LayoutDashboard, ExternalLink, Mountain, Layers, Satellite, Heart } from '@lucide/svelte';

	let { show = false, onClose }: { show?: boolean; onClose?: () => void } = $props();
	let manualTab = $state('intro');
</script>

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-overlay show" onclick={(e) => { if(e.target === e.currentTarget) onClose?.(); }}>
		<div class="modal-content-box">
			<button class="modal-close-btn" onclick={onClose}><X size={20} /></button>
			<div style="padding: 30px;">
				<h1 style="color: var(--text-primary); font-size: 24px; margin-bottom: 20px; text-align: center; font-weight: 800;">Manual del Sistema</h1>

				<div class="tabs-header" style="margin-bottom: 20px; justify-content: center; border-bottom: 1px solid var(--border-color);">
					<button class="tab-btn {manualTab === 'intro' ? 'active' : ''}" onclick={() => manualTab = 'intro'}>Introducción</button>
					<button class="tab-btn {manualTab === 'features' ? 'active' : ''}" onclick={() => manualTab = 'features'}>Funciones</button>
					<button class="tab-btn {manualTab === 'geospatial' ? 'active' : ''}" onclick={() => manualTab = 'geospatial'}>Contexto Geoespacial</button>
					<button class="tab-btn {manualTab === 'dashboard' ? 'active' : ''}" onclick={() => manualTab = 'dashboard'}>Dashboard</button>
					<button class="tab-btn {manualTab === 'credits' ? 'active' : ''}" onclick={() => manualTab = 'credits'}>Créditos</button>
					<button class="tab-btn {manualTab === 'donations' ? 'active' : ''}" onclick={() => manualTab = 'donations'}>Donaciones</button>
				</div>

				<div style="color: var(--text-primary); font-size: 15px; line-height: 1.6; padding-bottom: 20px; min-height: 250px;">
					{#if manualTab === 'intro'}
						<p style="margin-bottom: 16px;">SismoMonitor es un tablero analítico en tiempo real que integra datos sísmicos de <strong>USGS</strong>, <strong>EMSC</strong>, <strong>FUNVISIS</strong>, <strong>CSN</strong>, <strong>PRSN</strong>, <strong>UWI SRC</strong> e <strong>IPGP</strong>.</p>
						<h3 style="color: var(--text-primary); margin-top: 24px; margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Globe size={18} /> Modo Histórico</h3>
						<p style="margin-bottom: 16px;">Puedes seleccionar una fecha en el calendario superior. Al hacerlo, el feed en vivo se pausará y verás el historial de ese día en específico.</p>
					{/if}

					{#if manualTab === 'features'}
						<h3 style="color: var(--text-primary); margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><MapPin size={18} /> Feed de Sismos</h3>
						<ul style="list-style-type: disc; padding-left: 20px; margin-bottom: 16px;">
							<li><strong>Filtros:</strong> Puedes ver todos los sismos o filtrar exclusivamente por agencia: USGS, FUNVISIS, EMSC, CSN, PRSN, UWI o IPGP.</li>
							<li><strong>Ventana de tiempo:</strong> El feed en vivo muestra eventos del día actual (reset a medianoche). Para ver más días, usa el Modo Histórico.</li>
							<li><strong>Selección:</strong> Al hacer clic en un sismo, el mapa hará zoom automáticamente al epicentro y se resaltará el marcador.</li>
							<li><strong>Deduplicación:</strong> Si dos agencias reportan el mismo evento (menos de 5 minutos y 50km de diferencia), el sistema lo fusiona automáticamente.</li>
						</ul>

						<h3 style="color: var(--text-primary); margin-top: 24px; margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Radio size={18} /> Centro de Apoyo y Rescate (RRSS)</h3>
						<ul style="list-style-type: disc; padding-left: 20px; margin-bottom: 16px;">
							<li>Carrusel con reportes de inteligencia (OSINT) extraídos de canales de Telegram.</li>
							<li>Haz clic en una tarjeta para leer el mensaje completo. El carrusel se pausa mientras lees.</li>
						</ul>
					{/if}

					{#if manualTab === 'geospatial'}
						<h3 style="color: var(--text-primary); margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Mountain size={18} /> Volcanes y Alertas Volcánicas</h3>
						<p style="margin-bottom: 16px;">La capa de volcanes muestra 1.571 volcanes activos del Holoceno según el <strong>Smithsonian Global Volcanism Program</strong> (datos estáticos). Además, el sistema consulta cada 24 horas la API del <strong>USGS Volcano Hazards Program</strong> para resaltar volcanes con alertas activas (Advisory, Watch o Warning) en el banner superior.</p>

						<h3 style="color: var(--text-primary); margin-top: 24px; margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Layers size={18} /> Fallas Tectónicas y Límites de Placas</h3>
						<p style="margin-bottom: 16px;">Capa de líneas de falla y límites de placas tectónicas. Ayuda a interpretar la relación entre la ubicación de un sismo y las estructuras geológicas activas de la región.</p>

						<h3 style="color: var(--text-primary); margin-top: 24px; margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Satellite size={18} /> Daños por InSAR Sentinel-1 (Venezuela)</h3>
						<p style="margin-bottom: 16px;">Capa experimental de daños estimados por interferometría radar <strong>Sentinel-1</strong> (NASA / ESA), focalizada en Venezuela. Muestra estructuras afectadas detectadas tras eventos sísmicos recientes.</p>
					{/if}

					{#if manualTab === 'dashboard'}
						<h3 style="color: var(--text-primary); margin-bottom: 12px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><LayoutDashboard size={18} /> Secciones</h3>
						<ul style="list-style-type: disc; padding-left: 20px; margin-bottom: 16px;">
							<li style="margin-bottom: 8px;"><strong>Panel Izquierdo:</strong> Feed de sismos en vivo con filtros por agencia y navegación por páginas (Monitor, Análisis, Reportes, Configuración).</li>
							<li style="margin-bottom: 8px;"><strong>Mapa Central:</strong> Mapa interactivo con epicentros. Modo 2D (Leaflet) o Globo 3D (Mapbox GL). Toggle de capas: placas tectónicas, volcanes, daños InSAR.</li>
							<li style="margin-bottom: 8px;"><strong>Panel Derecho:</strong> Estadísticas rápidas, alertas volcánicas activas y eventos recientes con secciones colapsables.</li>
							<li style="margin-bottom: 8px;"><strong>Panel Inferior:</strong> Cuatro gráficos en tiempo real: línea temporal, distribución de magnitud, perfil de profundidad y energía acumulada.</li>
							<li style="margin-bottom: 8px;"><strong>Navegación Móvil:</strong> Bottom-nav con tres pestañas: Sismos, Mapa, Info/TV.</li>
						</ul>
					{/if}

					{#if manualTab === 'credits'}
						<h3 style="color: var(--text-primary); margin-bottom: 16px; font-weight: 600;">Desarrollo y Plataforma</h3>
						<ul style="list-style-type: none; padding-left: 0; margin-bottom: 24px;">
							<li>
								<strong style="color: var(--text-primary); font-size: 16px;">JAG-MEDIA SERVICIOS, C.A.</strong><br>
								<a href="https://jagmedia.com.ve" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> jagmedia.com.ve</a>
							</li>
						</ul>

						<h3 style="color: var(--text-primary); margin-bottom: 16px; font-weight: 600;">Instituciones y Fuentes de Datos</h3>
						<ul style="list-style-type: none; padding-left: 0; margin-bottom: 20px; display: flex; flex-direction: column; gap: 16px;">
							<li>
								<strong style="color: var(--text-primary);">FUNVISIS</strong> - Fundación Venezolana de Investigaciones Sismológicas<br>
								<a href="http://www.funvisis.gob.ve/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> funvisis.gob.ve</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">USGS</strong> - Servicio Geológico de los Estados Unidos<br>
								<a href="https://earthquake.usgs.gov/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> earthquake.usgs.gov</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">EMSC</strong> - Centro Sismológico Euromediterráneo<br>
								<a href="https://www.emsc-csem.org/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> emsc-csem.org</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">CSN</strong> - Centro Sismológico Nacional (Chile)<br>
								<a href="https://sismologia.cl/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> sismologia.cl</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">PRSN</strong> - Puerto Rico Seismic Network<br>
								<a href="http://www.prsn.uprm.edu/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> prsn.uprm.edu</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">UWI SRC</strong> - Seismic Research Centre (Caribe Oriental)<br>
								<a href="https://uwiseismic.com/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> uwiseismic.com</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">IPGP</strong> - Institut de Physique du Globe de Paris (Antillas Francesas)<br>
								<a href="https://ws.ipgp.fr/" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> ws.ipgp.fr</a>
							</li>
						</ul>
					{/if}
					{#if manualTab === 'donations'}
						<h3 style="color: var(--text-primary); margin-bottom: 16px; font-weight: 600; display: flex; align-items: center; gap: 6px;"><Heart size={18} /> Apoya SismoMonitor</h3>
						<p style="margin-bottom: 16px;">SismoMonitor es un proyecto independiente mantenido por <strong>JAG-MEDIA SERVICIOS, C.A.</strong>. Si te resulta útil, considera apoyar el desarrollo y mantenimiento de la infraestructura de datos en tiempo real.</p>
						<ul style="list-style-type: none; padding-left: 0; margin-bottom: 24px; display: flex; flex-direction: column; gap: 12px;">
							<li>
								<strong style="color: var(--text-primary);">PayPal</strong><br>
								<a href="https://jagmedia.com.ve" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> paypal.me/jagmedia</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">Ko-fi</strong><br>
								<a href="https://jagmedia.com.ve" target="_blank" style="color: var(--accent); text-decoration: none; font-size: 13px; display: inline-flex; align-items: center; gap: 4px; margin-top: 4px;"><ExternalLink size={12} /> ko-fi.com/jagmedia</a>
							</li>
							<li>
								<strong style="color: var(--text-primary);">Binance Pay / Cripto</strong><br>
								<span style="font-family: monospace; color: var(--text-secondary); font-size: 13px; margin-top: 4px; display: inline-block;">---</span>
							</li>
						</ul>
						<p style="font-size: 13px; color: var(--text-secondary);">Contacta directamente para coordinar apoyo empresarial o reportar fallos: <a href="mailto:contacto@jagmedia.com.ve" style="color: var(--accent); text-decoration: none;">contacto@jagmedia.com.ve</a></p>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: var(--bg-overlay);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		z-index: 9999;
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 20px;
		opacity: 0;
		pointer-events: none;
		transition: opacity 0.3s ease;
	}
	.modal-overlay.show {
		opacity: 1;
		pointer-events: auto;
	}
	.modal-content-box {
		position: relative;
		background: var(--bg-darker);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		width: 90vw;
		max-width: 900px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.8);
		animation: slideUp 0.3s ease;
	}
	.modal-close-btn {
		position: absolute;
		top: 16px;
		right: 16px;
		background: var(--bg-input);
		border: none;
		color: var(--text-primary);
		width: 32px;
		height: 32px;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
		z-index: 1;
	}
	.modal-close-btn:hover { background: #ff4444; }
	@keyframes slideUp {
		from { opacity: 0; transform: translateY(20px); }
		to { opacity: 1; transform: translateY(0); }
	}
</style>
