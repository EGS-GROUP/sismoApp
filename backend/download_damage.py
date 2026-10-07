#!/usr/bin/env python3
"""
Descarga completa del dataset Sentinel-1 Damage desde ArcGIS FeatureServer.
Usa paginación por fid (objectId) que es más confiable que resultOffset.
"""
import urllib.request
import urllib.parse
import json
import sys
import time

BASE_URL = "https://services7.arcgis.com/WSiUmUhlFx4CtMBB/ArcGIS/rest/services/202610_s1_likelydmgareas/FeatureServer/0/query"
PAGE_SIZE = 1000  # Más pequeño para evitar timeouts
OUTPUT_FILE = "/var/www/html/sismoApp/frontend/static/damage_sentinel1.json"

def fetch_page(where_clause, retry=3):
    """Fetch a single page with retries."""
    params = urllib.parse.urlencode({
        "where": where_clause,
        "outFields": "fid,damage_probability",
        "returnGeometry": "true",
        "outSR": "4326",
        "orderByFields": "fid ASC",
        "resultRecordCount": str(PAGE_SIZE),
        "f": "geojson"
    })
    url = f"{BASE_URL}?{params}"
    
    for attempt in range(retry):
        try:
            req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0"})
            with urllib.request.urlopen(req, timeout=90) as resp:
                data = json.loads(resp.read().decode("utf-8"))
            if "error" in data:
                raise Exception(f"ArcGIS error: {data['error'].get('message','unknown')}")
            return data.get("features", [])
        except Exception as e:
            print(f"  Reintento {attempt+1}/{retry}: {e}", file=sys.stderr)
            time.sleep(3 * (attempt + 1))
    return None

all_features = []
last_fid = -1
page_num = 0
consecutive_empty = 0

while True:
    # Paginación por fid: pide features con fid > último fid visto
    where = f"damage=1 AND fid>{last_fid}"
    
    features = fetch_page(where)
    if features is None:
        print(f"FALLO FATAL en página {page_num} (fid>{last_fid})", file=sys.stderr)
        # Guardar lo que tengamos hasta ahora
        break
    
    count = len(features)
    if count == 0:
        consecutive_empty += 1
        if consecutive_empty >= 2:
            break
        # Intentar saltar un rango de fids
        last_fid += PAGE_SIZE
        continue
    
    consecutive_empty = 0
    all_features.extend(features)
    
    # Obtener el fid máximo de esta página para la siguiente iteración
    max_fid = max(f["properties"]["fid"] for f in features)
    last_fid = max_fid
    page_num += 1
    
    print(f"Página {page_num}: {count} features, fid hasta {max_fid} (total: {len(all_features)})")
    
    if count < PAGE_SIZE:
        break
    
    time.sleep(0.3)

# Construir GeoJSON final
geojson = {
    "type": "FeatureCollection",
    "features": all_features
}

with open(OUTPUT_FILE, "w") as f:
    json.dump(geojson, f)

file_size_mb = len(json.dumps(geojson)) / (1024 * 1024)
print(f"\n=== COMPLETADO ===")
print(f"Total features descargados: {len(all_features)}")
print(f"Archivo: {OUTPUT_FILE}")
print(f"Tamaño aprox: {file_size_mb:.1f} MB")
