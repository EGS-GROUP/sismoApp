const usStates = [
	'AL', 'AK', 'AZ', 'AR', 'CA', 'CO', 'CT', 'DE', 'FL', 'GA', 'HI', 'ID', 'IL', 'IN', 'IA', 'KS', 'KY', 'LA', 'ME', 'MD', 'MA', 'MI', 'MN', 'MS', 'MO', 'MT', 'NE', 'NV', 'NH', 'NJ', 'NM', 'NY', 'NC', 'ND', 'OH', 'OK', 'OR', 'PA', 'RI', 'SC', 'SD', 'TN', 'TX', 'UT', 'VT', 'VA', 'WA', 'WV', 'WI', 'WY',
	'Alabama', 'Alaska', 'Arizona', 'Arkansas', 'California', 'Colorado', 'Connecticut', 'Delaware', 'Florida', 'Georgia', 'Hawaii', 'Idaho', 'Illinois', 'Indiana', 'Iowa', 'Kansas', 'Kentucky', 'Louisiana', 'Maine', 'Maryland', 'Massachusetts', 'Michigan', 'Minnesota', 'Mississippi', 'Missouri', 'Montana', 'Nebraska', 'Nevada', 'New Hampshire', 'New Jersey', 'New Mexico', 'New York', 'North Carolina', 'North Dakota', 'Ohio', 'Oklahoma', 'Oregon', 'Pennsylvania', 'Rhode Island', 'South Carolina', 'South Dakota', 'Tennessee', 'Texas', 'Utah', 'Vermont', 'Virginia', 'Washington', 'West Virginia', 'Wisconsin', 'Wyoming',
	'Puerto Rico', 'U.S. Virgin Islands', 'Guam', 'American Samoa', 'Northern Mariana Islands', 'PR', 'VI', 'GU', 'AS', 'MP'
];

export function extractRegion(locationString: string) {
	if (!locationString) return 'Desconocido';

	const locLower = locationString.toLowerCase();
	if (locLower.includes('venezuela')) return 'Venezuela';
	if (locLower.includes('mexico') || locLower.includes('méxico') || locLower.includes(', mx')) return 'México';
	if (locLower.includes('chile')) return 'Chile';
	if (locLower.includes('peru') || locLower.includes('perú')) return 'Perú';
	if (locLower.includes('japan') || locLower.includes('japón')) return 'Japón';
	if (locLower.includes('argentina')) return 'Argentina';
	if (locLower.includes('colombia')) return 'Colombia';
	if (locLower.includes('ecuador')) return 'Ecuador';
	if (locLower.includes('panama') || locLower.includes('panamá')) return 'Panamá';

	const parts = locationString.split(',');
	const region = parts.length > 1 ? parts[parts.length - 1].trim() : locationString.trim();

	if (usStates.some(s => s.toLowerCase() === region.toLowerCase())) {
		return 'Estados Unidos';
	}

	return region.split(' ')
		.filter(w => w.length > 0 && w.toLowerCase() !== 'region')
		.map(w => w.charAt(0).toUpperCase() + w.substring(1).toLowerCase())
		.join(' ').replace(/off the coast of /i, '').replace(/near the coast of /i, '').trim();
}

export function extractVenezuelaState(locationString: string) {
	if (!locationString) return 'Desconocido';
	const loc = locationString.toLowerCase();

	if (loc.includes('sucre') || loc.includes('carúpano') || loc.includes('carupano') || loc.includes('güiria') || loc.includes('guiria') || loc.includes('cumana') || loc.includes('cumaná')) return 'Sucre';
	if (loc.includes('carabobo') || loc.includes('morón') || loc.includes('moron') || loc.includes('puerto cabello') || loc.includes('valencia')) return 'Carabobo';
	if (loc.includes('miranda') || loc.includes('guatire') || loc.includes('guarenas') || loc.includes('los teques') || loc.includes('higuerote')) return 'Miranda';
	if (loc.includes('lara') || loc.includes('el tocuyo') || loc.includes('barquisimeto') || loc.includes('carora')) return 'Lara';
	if (loc.includes('mérida') || loc.includes('merida') || loc.includes('el vigía') || loc.includes('el vigia') || loc.includes('tovar')) return 'Mérida';
	if (loc.includes('yaracuy') || loc.includes('san felipe') || loc.includes('chivacoa') || loc.includes('yaritagua')) return 'Yaracuy';
	if (loc.includes('aragua') || loc.includes('maracay') || loc.includes('el limón') || loc.includes('el limon') || loc.includes('cagua')) return 'Aragua';
	if (loc.includes('falcón') || loc.includes('falcon') || loc.includes('coro') || loc.includes('punto fijo') || loc.includes('boca de aroa')) return 'Falcón';
	if (loc.includes('zulia') || loc.includes('maracaibo') || loc.includes('lagunillas') || loc.includes('cabimas')) return 'Zulia';
	if (loc.includes('trujillo') || loc.includes('boconó') || loc.includes('bocono') || loc.includes('valera') || loc.includes('pampanito')) return 'Trujillo';
	if (loc.includes('vargas') || loc.includes('la guaira') || loc.includes('catia la mar') || loc.includes('naiguata') || loc.includes('naiguatá')) return 'La Guaira';
	if (loc.includes('táchira') || loc.includes('tachira') || loc.includes('san cristóbal') || loc.includes('san cristobal') || loc.includes('la grita')) return 'Táchira';
	if (loc.includes('anzoátegui') || loc.includes('anzoategui') || loc.includes('puerto la cruz') || loc.includes('barcelona') || loc.includes('el tigre')) return 'Anzoátegui';
	if (loc.includes('monagas') || loc.includes('maturín') || loc.includes('maturin') || loc.includes('caripe')) return 'Monagas';
	if (loc.includes('nueva esparta') || loc.includes('margarita') || loc.includes('porlamar') || loc.includes('la asunción') || loc.includes('la asuncion')) return 'Nueva Esparta';
	if (loc.includes('caracas') || loc.includes('distrito capital') || loc.includes('libertador') || loc.includes('chacao') || loc.includes('baruta')) return 'Distrito Capital';
	if (loc.includes('apure') || loc.includes('san fernando') || loc.includes('guasdualito')) return 'Apure';
	if (loc.includes('barinas') || loc.includes('socopó') || loc.includes('socopo')) return 'Barinas';
	if (loc.includes('bolívar') || loc.includes('bolivar') || loc.includes('ciudad guayana') || loc.includes('puerto ordaz')) return 'Bolívar';
	if (loc.includes('cojedes') || loc.includes('san carlos') || loc.includes('tinaquillo')) return 'Cojedes';
	if (loc.includes('delta amacuro') || loc.includes('tucupita')) return 'Delta Amacuro';
	if (loc.includes('guárico') || loc.includes('guarico') || loc.includes('san juan')) return 'Guárico';
	if (loc.includes('portuguesa') || loc.includes('guanare') || loc.includes('acarigua')) return 'Portuguesa';
	if (loc.includes('amazonas') || loc.includes('puerto ayacucho')) return 'Amazonas';
	return 'Otras Regiones';
}
