import sqlite3
import time
import datetime

conn = sqlite3.connect('sismos.db')
c = conn.cursor()

now = int(time.time() * 1000)
yesterday = now - (24 * 60 * 60 * 1000)

c.execute("SELECT id, location, datetime(time/1000, 'unixepoch', 'localtime') FROM funvisis_history WHERE time > ?", (yesterday,))
for row in c.fetchall():
    print(row)

today_noon = datetime.datetime.now().replace(hour=11, minute=0, second=0, microsecond=0)
noon_ms = int(today_noon.timestamp() * 1000)

c.execute("DELETE FROM funvisis_history WHERE time > ?", (noon_ms,))
print('Deleted', c.rowcount, 'bad records.')

conn.commit()
conn.close()
