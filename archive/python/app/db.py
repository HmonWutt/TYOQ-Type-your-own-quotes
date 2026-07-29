import sqlite3
import os
import pprint

def main():
    con = sqlite3.connect("data/seed.db")
    cursor = con.cursor()
    print ("Connected!\n")
    cursor.execute("SELECT * FROM quotes where word_count < 50 limit 5")
    records = cursor.fetchall()
    quotes = []
    for record in records:
        quotes.append(record[1])

    return quotes

if __name__ == "__main__":
	main()
