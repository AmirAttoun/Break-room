# Break-Room

## What is Break-Room.ch
The [Break-Room](https://break-room.ch) website is a simple solution to a problem we faced at our school. The problem was that it was very difficult to find unoccupied rooms to work in quietly. Students often had to open each room to check if it was free, disrupting the classes inside. My solution to this was a website that always shows the empty rooms at any given time.
The site is currently developed and maintained by me, with financial support from KZU.
  
## Additional Features
In addition to displaying empty rooms, I have added several other features.

**Displaying consecutive empty rooms:**
One of these features allows you to see which rooms are empty for two consecutive lessons. This is useful when you need a space to study for more than just one lesson.

**Displaying canceled lessons:**
The website also lets you display all rooms where lessons have been canceled.

  
## Break-Room.ch at your school
Break-Room.ch could make life easier for students and staff at your school, too. With a simple interface and real-time data, it eliminates the guesswork of finding available study spaces. No more interrupting classes or wasting time hunting for an empty room. It’s a straightforward, effective tool that creates a quieter, more focused learning environment. Plus, additional features like tracking consecutive free periods and canceled lessons make it even more convenient. If your school faces similar challenges, Break-Room.ch is the solution you’ve been looking for.

## Technical
**Backend**
The backend uses a GO-lang script to fetch the necessary lesson data from the official school website. This data is updated twice a week to ensure it’s always current. Each room’s availability is checked against a complete list of all school rooms to determine which are unoccupied during each lesson. The data is delivered to users via a simple FastAPI application. If you’d like more in-depth information about the site, you can check out the code on my [github](https://github.com/HelloAlex4/Break-room).
