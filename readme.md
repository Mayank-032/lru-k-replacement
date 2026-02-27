**Buffer POOL Manager**

Introduction:
- The buffer pool is responsible for moving physical pages of data back and forth from buffers in main memory to persistent storage. 

- It also behaves as a cache
    Key-Eviction Policy: keeping frequently used pages in memory for faster access, and evicting unused or cold pages back out to storage. Similar to LRU.

- In addition to behaving as a cache, the buffer pool manager allows a DBMS to support databases that are larger than the amount of memory available to the system. Consider a computer with 1 GB of memory (RAM). If we want to manage a 2 GB database, a buffer pool manager gives us the ability to interact with this database without needing to fit its entire contents in memory.

- The I/O operations that the buffer pool executes are abstracted away from other parts of the DBMS. Here, The buffer pool manager does not need to understand the contents of these pages, it only needs to know where the data is located.

Implementation
- 







k - 2
Cache: [
    1 - 1, 4 -> 4, 12 
    2 - 2, 5 -> 5, 6 -> 6, 9
    3 - 3, 7 -> 7, 8 -> 8, 10 -> 10, 11

    evict(1) and now cache looks like
    2 - 2, 5 -> 5, 6 -> 6, 9
    3 - 3, 7 -> 7, 8 -> 8, 10 -> 10, 11
    4 - 12, 13
]
Size - 3

1(S),
2(S),
3(S),
1(G),
2(G),
2(S),
3(G),
3(S),
2(G),
3(G),
3(G),
1(G),
4(S),
4(G)