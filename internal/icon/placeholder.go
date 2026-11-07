package icon

// Placeholder reports that the icon is still the skeleton's stand-in. It is
// false now that the game's own icon is drawn (assets/sprites/icon.txt);
// while it is true, cmd/winicon refuses to write the exe's icon, so a
// release can't be built with the stand-in.
const Placeholder = false
