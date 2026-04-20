package sqlite

import (
	"encoding/binary"
	"math"

	"github.com/david22573/GoRadio/app/types"
)

// SerializeFloat32 converts []float32 to little-endian bytes
func SerializeFloat32(v []float32) []byte {
	b := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

// DeserializeFloat32 converts little-endian bytes to []float32
func DeserializeFloat32(b []byte) []float32 {
	v := make([]float32, len(b)/4)
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

// InsertTrackVector adds or updates a track embedding
func (c *Client) InsertTrackVector(trackID uint, embedding []float64) error {
	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	query := `INSERT INTO track_vectors(track_id, embedding) VALUES(?, ?)
              ON CONFLICT(track_id) DO UPDATE SET embedding=excluded.embedding`
	_, err := c.db.Exec(query, trackID, blob)
	return err
}

// DistanceMetric defines the distance calculation method
type DistanceMetric string

const (
	DistanceL2     DistanceMetric = "distance_l2"
	DistanceCosine DistanceMetric = "distance_cosine"
)

// GetTrackVector retrieves the embedding for a track
func (c *Client) GetTrackVector(trackID uint) ([]float64, error) {
	var blob []byte
	query := "SELECT embedding FROM track_vectors WHERE track_id = ?"
	err := c.db.QueryRow(query, trackID).Scan(&blob)
	if err != nil {
		return nil, err
	}
	
	f32s := DeserializeFloat32(blob)
	f64s := make([]float64, len(f32s))
	for i, v := range f32s {
		f64s[i] = float64(v)
	}
	return f64s, nil
}

// FindNearestNeighbors finds K nearest neighbors for a given trackID
func (c *Client) FindNearestNeighbors(trackID uint, k int, excludeIDs []uint) ([]types.Track, []float64, error) {
	embedding, err := c.GetTrackVector(trackID)
	if err != nil {
		return nil, nil, err
	}

	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	if k <= 0 {
		k = 10
	}

	query := `
		SELECT track_id, vec_distance_l2(embedding, ?) as distance
		FROM track_vectors
		WHERE track_id != ?
		ORDER BY distance
		LIMIT ?`
	
	rows, err := c.db.Query(query, blob, trackID, k+len(excludeIDs))
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	excludeMap := make(map[uint]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	var tracks []types.Track
	var distances []float64
	for rows.Next() {
		var id uint
		var dist float64
		if err := rows.Scan(&id, &dist); err != nil {
			return nil, nil, err
		}
		if excludeMap[id] {
			continue
		}
		track, err := c.GetTrackByID(id)
		if err == nil {
			tracks = append(tracks, *track)
			distances = append(distances, dist)
		}
		if len(tracks) >= k {
			break
		}
	}
	return tracks, distances, nil
}

// FindDistantNeighbors finds tracks within distance range [minDist, maxDist]
func (c *Client) FindDistantNeighbors(trackID uint, minDist, maxDist float64, k int) ([]types.Track, error) {
	embedding, err := c.GetTrackVector(trackID)
	if err != nil {
		return nil, err
	}

	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	query := `
		SELECT track_id
		FROM (
			SELECT track_id, vec_distance_l2(embedding, ?) AS dist
			FROM track_vectors
			WHERE track_id != ?
			ORDER BY dist
		)
		WHERE dist >= ? AND dist <= ?
		LIMIT ?`

	rows, err := c.db.Query(query, blob, trackID, minDist, maxDist, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []types.Track
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		track, err := c.GetTrackByID(id)
		if err == nil {
			tracks = append(tracks, *track)
		}
	}
	return tracks, nil
}

// SearchRange finds tracks within distance range [minDist, maxDist]
func (c *Client) SearchRange(embedding []float64, minDist, maxDist float64, k int) ([]uint, error) {
	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	query := `
		SELECT track_id
		FROM (
			SELECT track_id, vec_distance_l2(embedding, ?) AS dist
			FROM track_vectors
		)
		WHERE dist >= ? AND dist <= ?
		LIMIT ?`

	rows, err := c.db.Query(query, blob, minDist, maxDist, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Deprecated: Use GetTrackVector
func (c *Client) GetVectorByID(trackID uint) ([]float64, error) {
	return c.GetTrackVector(trackID)
}

// Deprecated: Use FindDistantNeighbors-like query if needed, or keeping it for compatibility
func (c *Client) GetDistantTracks(embedding []float64, k int) ([]types.Track, error) {
	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	query := `
		SELECT track_id
		FROM track_vectors
		ORDER BY vec_distance_l2(embedding, ?) DESC
		LIMIT ?`
	
	rows, err := c.db.Query(query, blob, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []types.Track
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		track, err := c.GetTrackByID(id)
		if err == nil {
			tracks = append(tracks, *track)
		}
	}
	return tracks, nil
}

// SearchKNN finds K nearest neighbors for a given embedding
func (c *Client) SearchKNN(embedding []float64, k int, metric DistanceMetric) ([]uint, error) {
	ids, _, err := c.SearchKNNWithDistances(embedding, k, metric)
	return ids, err
}

// SearchKNNWithDistances finds K nearest neighbors for a given embedding and returns their IDs and distances
func (c *Client) SearchKNNWithDistances(embedding []float64, k int, metric DistanceMetric) ([]uint, []float64, error) {
	f32Embedding := make([]float32, len(embedding))
	for i, v := range embedding {
		f32Embedding[i] = float32(v)
	}
	blob := SerializeFloat32(f32Embedding)

	if k <= 0 {
		k = 10
	}

	var query string
	switch metric {
	case DistanceCosine:
		query = `SELECT track_id, vec_distance_cosine(embedding, ?) as distance FROM track_vectors ORDER BY distance LIMIT ?`
	default:
		query = `SELECT track_id, vec_distance_l2(embedding, ?) as distance FROM track_vectors ORDER BY distance LIMIT ?`
	}

	rows, err := c.db.Query(query, blob, k)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids []uint
	var distances []float64
	for rows.Next() {
		var id uint
		var dist float64
		if err := rows.Scan(&id, &dist); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		distances = append(distances, dist)
	}
	return ids, distances, nil
}
