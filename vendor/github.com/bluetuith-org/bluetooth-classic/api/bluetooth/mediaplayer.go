package bluetooth

// MediaPlayer describes a function call interface to invoke media player/control
// related functions on a device.
type MediaPlayer interface {
	AudioProfiles() ([]AudioProfile, error)
	SetAudioProfile(profile AudioProfile) error

	Properties() (MediaData, error)

	Play() error
	Pause() error
	TogglePlayPause() error

	Next() error
	Previous() error
	FastForward() error
	Rewind() error

	Stop() error
}

// AudioProfile stores the device's audio profile information.
type AudioProfile struct {
	// Name holds the name of the audio profile.
	Name string `json:"name,omitempty" codec:"Name,omitempty" doc:"Holds the name of the audio profile."`

	// Description holds a brief description of the audio profile.
	Description string `json:"description,omitempty" codec:"Description,omitempty" doc:"Holds a brief description of the audio profile."`

	// Index holds the current position of the audio profile in the list.
	Index uint32 `json:"index,omitempty" codec:"Index,omitempty" doc:"Holds the current position of the audio profile in the list."`

	// Active specifies if the profile is active.
	Active bool `json:"active,omitempty" codec:"Active,omitempty" doc:"Specifies if the profile is active."`
}

// MediaStatus indicates the status of the media player.
type MediaStatus string

// The different values for the media player status.
const (
	MediaPlaying     MediaStatus = "playing"
	MediaPaused      MediaStatus = "paused"
	MediaForwardSeek MediaStatus = "forward-seek"
	MediaReverseSeek MediaStatus = "reverse-seek"
	MediaStopped     MediaStatus = "stopped"
)

// MediaData holds the media player information.
type MediaData struct {
	// Address holds the Bluetooth MAC address of the device.
	Address MacAddress `json:"address,omitempty" codec:"Address,omitempty" doc:"The Bluetooth MAC address of the device."`

	// Status indicates the status of the player.
	Status MediaStatus `json:"status,omitempty" codec:"Status,omitempty" enum:"playing,paused,forward-seek,reverse-seek,stopped" doc:"Indicates the status of the player."`

	// Position indicates the current position of the playing track.
	Position uint32 `json:"position,omitempty" codec:"Position,omitempty" doc:"Indicates the current position of the playing track."`

	TrackData
}

// TrackData describes the track properties of the currently playing media.
type TrackData struct {
	// Title holds the title name of the track.
	Title string `json:"title,omitempty" codec:"Title,omitempty" doc:"The title name of the track."`

	// Album holds the album name of the track.
	Album string `json:"album,omitempty" codec:"Album,omitempty" doc:"The album name of the track."`

	// Artist holds the artist name of the track.
	Artist string `json:"artist,omitempty" codec:"Artist,omitempty" doc:"The artist name of the track."`

	// Duration holds the duration of the track.
	Duration uint32 `json:"duration,omitempty" codec:"Duration,omitempty" doc:"The duration of the track."`

	// TrackNumber holds the playlist position of the track.
	TrackNumber uint32 `json:"track_number,omitempty" codec:"TrackNumber,omitempty" doc:"The playlist position of the track."`

	// TotalTracks holds the total number of tracks.
	TotalTracks uint32 `json:"total_tracks,omitempty" codec:"TotalTracks,omitempty" doc:"The total number of tracks."`
}
