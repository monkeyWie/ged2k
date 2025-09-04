package jed2k

// Constants defines the core constants used throughout the ed2k protocol implementation
const (
	PieceSize         int64 = 9728000         // PIECE_SIZE = 9728000L
	BlockSize         int64 = 190 * 1024      // BLOCK_SIZE = 190kb = PIECE_SIZE/50  
	BlockSizeInt      int   = int(BlockSize)  // BLOCK_SIZE_INT
	BlocksPerPiece    int   = int(PieceSize / BlockSize) // 50 blocks per piece
	HighestLowidEd2k  int64 = 16777216        // HIGHEST_LOWID_ED2K
	RequestQueueSize  int   = 3               // REQUEST_QUEUE_SIZE
	PartsInRequest    int   = 3               // PARTS_IN_REQUEST
)