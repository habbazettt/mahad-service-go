package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type QTable map[string]map[string]float64

type HistoricalInfo struct {
	Jadwal            string  `json:"-"`
	PersentaseEfektif float64 `json:"Persentase Efektif (%)"`
}

var (
	QTableModel    QTable
	HistoricalBest []HistoricalInfo
)

func LoadQlearningModels() error {
	qTableFile, err := os.ReadFile("./q_table_model.json")
	if err != nil {
		return fmt.Errorf("gagal membaca q_table_model.json: %w", err)
	}
	if err := json.Unmarshal(qTableFile, &QTableModel); err != nil {
		return fmt.Errorf("gagal unmarshal q_table_model.json: %w", err)
	}
	fmt.Println("Model Q-Table berhasil dimuat.")

	historicalFile, err := os.ReadFile("./historical_best.json")
	if err != nil {
		fmt.Println("Peringatan: file historical_best.json tidak ditemukan. Fitur fallback historis tidak akan aktif.")
	} else {
		var tempHistoricalMap map[string]HistoricalInfo
		if err := json.Unmarshal(historicalFile, &tempHistoricalMap); err != nil {
			return fmt.Errorf("gagal unmarshal historical_best.json: %w", err)
		}

		for jadwal, info := range tempHistoricalMap {
			info.Jadwal = jadwal
			HistoricalBest = append(HistoricalBest, info)
		}

		sort.Slice(HistoricalBest, func(i, j int) bool {
			return HistoricalBest[i].PersentaseEfektif > HistoricalBest[j].PersentaseEfektif
		})
		fmt.Println("Data historis untuk fallback berhasil dimuat dan diurutkan.")
	}

	return nil
}
