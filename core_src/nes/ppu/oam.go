package ppu

type ObjAttribEntry struct {
    Y uint8     // Y position 
    ID uint8    // ID of tile from pattern memory
    Attribute   uint8   // how sprite should be rendered
    X uint8     // X position
}

