import psycopg2
from dotenv import load_dotenv
import os
import pprint

def main():
	#Define our connection string
    load_dotenv()
    POSTGRES_USER=os.getenv('POSTGRES_USER')
    POSTGRES_PASSWORD=os.getenv('POSTGRES_PASSWORD')
    POSTGRES_DB=os.getenv('POSTGRES_DB')
    HOST=os.getenv('HOST')
    PORT=os.getenv('PORT')
    conn_string = f"host={HOST} dbname={POSTGRES_DB} user={POSTGRES_USER} password={POSTGRES_PASSWORD}"

	# print the connection string we will use to connect
    print (f"Connecting to database...")

	# get a connection, if a connect cannot be made an exception will be raised here
    conn = psycopg2.connect(conn_string)

	# conn.cursor will return a cursor object, you can use this cursor to perform queries
    cursor = conn.cursor()
    print ("Connected!\n")
    cursor.execute("SELECT * FROM quotes WHERE word_count > 50 AND word_count < 70")
    records = cursor.fetchall()
    quotes = []
    for record in records:
        quotes.append(record[1])

    return quotes

if __name__ == "__main__":
	main()
