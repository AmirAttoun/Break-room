from fastapi import FastAPI, HTTPException, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
import aiosqlite
import asyncio
import time
from datetime import datetime, timedelta

app = FastAPI()

# Dependency to get a new database connection per request
async def get_db():
    try:
        db = await aiosqlite.connect('/root/go/src/brakeRoom/brake-room.db', isolation_level=None)
        return db
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Database connection failed: {str(e)}")

@app.get("//kzu")
async def get_data(db: aiosqlite.Connection = Depends(get_db)):
    current_time_str = time.strftime('%Y-%m-%d %H:%M:%S')

    # Convert the string to a datetime object
    current_time = datetime.strptime(current_time_str, '%Y-%m-%d %H:%M:%S')

    # Add 15 minutes
    new_time = current_time + timedelta(minutes=15)

    # Convert the new time back to a string
    current_time = new_time.strftime('%Y-%m-%d %H:%M:%S')


    print(f"Current time: {current_time}")

    try:
        query =  """
            WITH LatestEvent AS (
                SELECT id, startTime
                FROM times
                WHERE startTime < ?
                ORDER BY startTime DESC
                LIMIT 1
            )

            SELECT rooms.room, times.startTime, rooms.consecutive, rooms.canceled
            FROM rooms
            JOIN times ON rooms.id = times.id
            WHERE rooms.id = (SELECT id FROM LatestEvent);
        """
        cursor = await db.execute(query, (current_time,))
        rows = await cursor.fetchall()
        columns = [column[0] for column in cursor.description]
        await cursor.close()
        return [dict(zip(columns, row)) for row in rows]
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query execution failed: {str(e)}")
    finally:
        await db.close()

@app.post("//error-logging")
async def log_error(request: Request, db: aiosqlite.Connection = Depends(get_db)):
    try:
        data = await request.json()
        error_message = data.get('error', 'No error message provided')
        
        query = "INSERT INTO errors (error) VALUES (?);"
        await db.execute(query, (error_message,))
        await db.commit()

        return {"status": "success", "message": "Error logged successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to log error: {str(e)}")
    finally:
        await db.close()


@app.get("//test-db")
async def test_db(db: aiosqlite.Connection = Depends(get_db)):
    try:
        # Just fetch some data to see if the connection and queries are working
        cursor = await db.execute("SELECT * FROM times LIMIT 5;")
        rows = await cursor.fetchall()
        columns = [column[0] for column in cursor.description]
        await cursor.close()

        return [dict(zip(columns, row)) for row in rows]
    except Exception as e:
        return {"error": str(e)}
    finally:
        await db.close()

# Start the FastAPI application using Uvicorn
if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
