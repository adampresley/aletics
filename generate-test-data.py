#!/usr/bin/env python3
"""
Generate fake test data for aletics properties for the last 30 days.
Reads DSN from .env in the current directory and inserts into SQLite or Postgres.

Usage:
    python generate-test-data.py [--days 30] [--clear]

Options:
    --days N    Number of days back to generate data for (default: 30)
    --clear     Delete existing events for these property IDs before inserting
"""

import argparse
import os
import random
import re
import sqlite3
import sys
from datetime import datetime, timedelta, timezone

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

PROPERTY_IDS = [1, 4, 7]

PATHS = [
    ("/", 20),                          # (path, relative weight)
    ("/index.html", 15),
    ("/about.html", 8),
    ("/features.html", 12),
    ("/pricing.html", 10),
    ("/blog.html", 8),
    ("/blog/getting-started.html", 6),
    ("/docs.html", 7),
    ("/signup.html", 9),
    ("/contact.html", 5),
]

QUERY_STRINGS = [
    ("", 60),                           # (value, relative weight)
    ("?ref=github", 8),
    ("?ref=twitter", 6),
    ("?utm_source=google", 7),
    ("?utm_source=newsletter", 5),
    ("?utm_campaign=launch", 6),
    ("?page=2", 4),
    ("?utm_campaign=launch", 4),
]

BROWSERS = [
    ("Chrome", 55),
    ("Safari", 25),
    ("Firefox", 12),
    ("Edge", 8),
]

# country_code -> (country_name, continent_name, continent_code)
COUNTRIES = {
    "US": ("United States", "North America", "NA"),
    "GB": ("United Kingdom", "Europe", "EU"),
    "CA": ("Canada", "North America", "NA"),
    "DE": ("Germany", "Europe", "EU"),
    "FR": ("France", "Europe", "EU"),
    "AU": ("Australia", "Oceania", "OC"),
    "JP": ("Japan", "Asia", "AS"),
}

COUNTRY_WEIGHTS = [55, 10, 8, 7, 6, 5, 4]   # matches COUNTRIES order above

# Rough events-per-day range for each property
EVENTS_PER_DAY = {
    1: (8, 25),
    4: (5, 18),
    7: (3, 12),
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def weighted_choice(choices):
    """Pick a value from [(value, weight), ...] according to weights."""
    values, weights = zip(*choices)
    return random.choices(values, weights=weights, k=1)[0]


def random_timestamp(day: datetime) -> datetime:
    """Return a random UTC datetime within the given day."""
    hour = random.randint(0, 23)
    minute = random.randint(0, 59)
    second = random.randint(0, 59)
    return day.replace(hour=hour, minute=minute, second=second, tzinfo=timezone.utc)


def load_env(path=".env"):
    """Load key=value pairs from a .env file."""
    env = {}
    if not os.path.exists(path):
        return env
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            m = re.match(r'^([A-Z_]+)\s*=\s*"?([^"]*)"?$', line)
            if m:
                env[m.group(1)] = m.group(2)
    return env


def generate_events(num_days: int):
    """Yield (property_id, ts, path, query_string, browser, country_code, country, continent, continent_code) tuples."""
    today = datetime.now(timezone.utc).replace(hour=0, minute=0, second=0, microsecond=0)
    country_items = list(COUNTRIES.items())  # [(code, (name, continent, cont_code)), ...]

    for property_id in PROPERTY_IDS:
        min_ev, max_ev = EVENTS_PER_DAY[property_id]
        for day_offset in range(num_days):
            day = today - timedelta(days=day_offset)
            count = random.randint(min_ev, max_ev)
            for _ in range(count):
                ts = random_timestamp(day)
                path = weighted_choice(PATHS)
                qs = weighted_choice(QUERY_STRINGS)
                browser = weighted_choice(BROWSERS)
                country_code, (country, continent, continent_code) = random.choices(
                    country_items, weights=COUNTRY_WEIGHTS, k=1
                )[0]
                yield (property_id, ts, path, qs, browser, country_code, country, continent, continent_code)


# ---------------------------------------------------------------------------
# SQLite
# ---------------------------------------------------------------------------

def insert_sqlite(dsn: str, events, property_ids, clear: bool):
    # Strip "file:" prefix
    db_path = dsn.removeprefix("file:")
    conn = sqlite3.connect(db_path)
    cur = conn.cursor()

    if clear:
        placeholders = ",".join("?" * len(property_ids))
        cur.execute(f"DELETE FROM events WHERE property_id IN ({placeholders})", property_ids)
        print(f"Cleared existing events for properties {property_ids}")

    sql = """
        INSERT INTO events
            (created_at, updated_at, deleted_at, property_id, path, query_string,
             browser, country_code, country, continent, continent_code)
        VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)
    """
    rows = []
    for (prop_id, ts, path, qs, browser, country_code, country, continent, continent_code) in events:
        ts_str = ts.strftime("%Y-%m-%d %H:%M:%S")
        rows.append((ts_str, ts_str, prop_id, path, qs, browser, country_code, country, continent, continent_code))

    cur.executemany(sql, rows)
    conn.commit()
    conn.close()
    return len(rows)


# ---------------------------------------------------------------------------
# Postgres
# ---------------------------------------------------------------------------

def insert_postgres(dsn: str, events, property_ids, clear: bool):
    try:
        import psycopg2
    except ImportError:
        print("ERROR: psycopg2 is required for Postgres. Install it with: pip install psycopg2-binary")
        sys.exit(1)

    conn = psycopg2.connect(dsn)
    cur = conn.cursor()

    if clear:
        cur.execute("DELETE FROM events WHERE property_id = ANY(%s)", (list(property_ids),))
        print(f"Cleared existing events for properties {list(property_ids)}")

    sql = """
        INSERT INTO events
            (created_at, updated_at, deleted_at, property_id, path, query_string,
             browser, country_code, country, continent, continent_code)
        VALUES (%s, %s, NULL, %s, %s, %s, %s, %s, %s, %s, %s)
    """
    rows = []
    for (prop_id, ts, path, qs, browser, country_code, country, continent, continent_code) in events:
        rows.append((ts, ts, prop_id, path, qs, browser, country_code, country, continent, continent_code))

    cur.executemany(sql, rows)
    conn.commit()
    cur.close()
    conn.close()
    return len(rows)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--days", type=int, default=30, help="Number of days to generate (default: 30)")
    parser.add_argument("--clear", action="store_true", help="Delete existing events for these properties first")
    args = parser.parse_args()

    env = load_env()
    dsn = env.get("DSN", "file:./aletics.db")
    print(f"DSN: {dsn}")
    print(f"Generating {args.days} days of data for properties {PROPERTY_IDS}...")

    events = list(generate_events(args.days))

    if dsn.startswith("postgres"):
        count = insert_postgres(dsn, events, PROPERTY_IDS, args.clear)
    else:
        count = insert_sqlite(dsn, events, PROPERTY_IDS, args.clear)

    print(f"Inserted {count} events across {len(PROPERTY_IDS)} properties.")


if __name__ == "__main__":
    main()
