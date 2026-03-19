package dto

type ClusterRequest struct {
	Texts          []string `json:"texts"`
	MinClusterSize int      `json:"min_cluster_size"`
}

type ClusterTopicDTO struct {
	ClusterID      int      `json:"cluster_id"`
	Size           int      `json:"size"`
	Keywords       []string `json:"keywords"`
	Representative string   `json:"representative"`
	Messages       []string `json:"messages"`
}

type ClusterResponse struct {
	NumClusters     int               `json:"num_clusters"`
	NumNoise        int               `json:"num_noise"`
	SilhouetteScore float64           `json:"silhouette_score"`
	Clusters        []ClusterTopicDTO `json:"clusters"`
	Noise           []string          `json:"noise"`
}
