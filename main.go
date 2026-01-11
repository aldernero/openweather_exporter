package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Weather API response structures
type WeatherResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  float64 `json:"pressure"`
		Humidity  float64 `json:"humidity"`
		SeaLevel  float64 `json:"sea_level"`
		GrndLevel float64 `json:"grnd_level"`
	} `json:"main"`
	Visibility float64 `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All float64 `json:"all"`
	} `json:"clouds"`
	Dt  int64 `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

// Air Pollution API response structures
type AirPollutionResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	List []struct {
		Main struct {
			AQI int `json:"aqi"`
		} `json:"main"`
		Components struct {
			CO   float64 `json:"co"`
			NO   float64 `json:"no"`
			NO2  float64 `json:"no2"`
			O3   float64 `json:"o3"`
			SO2  float64 `json:"so2"`
			PM25 float64 `json:"pm2_5"`
			PM10 float64 `json:"pm10"`
			NH3  float64 `json:"nh3"`
		} `json:"components"`
		Dt int64 `json:"dt"`
	} `json:"list"`
}

// Prometheus metrics
var (
	// Weather metrics
	owWeatherTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_temp",
			Help: "Current temperature",
		},
		[]string{"station"},
	)
	owWeatherFeelsLike = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_feels_like",
			Help: "Feels like temperature",
		},
		[]string{"station"},
	)
	owWeatherTempMin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_temp_min",
			Help: "Minimum temperature",
		},
		[]string{"station"},
	)
	owWeatherTempMax = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_temp_max",
			Help: "Maximum temperature",
		},
		[]string{"station"},
	)
	owWeatherPressure = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_pressure",
			Help: "Atmospheric pressure in hPa",
		},
		[]string{"station"},
	)
	owWeatherHumidity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_humidity",
			Help: "Humidity percentage",
		},
		[]string{"station"},
	)
	owWeatherSeaLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_sea_level",
			Help: "Sea level pressure in hPa",
		},
		[]string{"station"},
	)
	owWeatherGrndLevel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_grnd_level",
			Help: "Ground level pressure in hPa",
		},
		[]string{"station"},
	)
	owWeatherVisibility = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_visibility",
			Help: "Visibility in meters",
		},
		[]string{"station"},
	)
	owWeatherWindSpeed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_wind_speed",
			Help: "Wind speed",
		},
		[]string{"station"},
	)
	owWeatherWindDeg = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_wind_deg",
			Help: "Wind direction in degrees",
		},
		[]string{"station"},
	)
	owWeatherClouds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_clouds",
			Help: "Cloud coverage percentage",
		},
		[]string{"station"},
	)
	owWeatherCondition = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_weather_condition",
			Help: "Weather condition ID",
		},
		[]string{"station", "main", "description"},
	)

	// Air pollution metrics
	owAirPollutionAQI = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_aqi",
			Help: "Air Quality Index (1-5)",
		},
		[]string{"station"},
	)
	owAirPollutionCO = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_co",
			Help: "Carbon monoxide concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionNO = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_no",
			Help: "Nitrogen monoxide concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionNO2 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_no2",
			Help: "Nitrogen dioxide concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionO3 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_o3",
			Help: "Ozone concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionSO2 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_so2",
			Help: "Sulphur dioxide concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionPM25 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_pm2_5",
			Help: "PM2.5 concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionPM10 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_pm10",
			Help: "PM10 concentration in μg/m³",
		},
		[]string{"station"},
	)
	owAirPollutionNH3 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ow_air_pollution_nh3",
			Help: "Ammonia concentration in μg/m³",
		},
		[]string{"station"},
	)
)

func init() {
	// Register weather metrics
	prometheus.MustRegister(owWeatherTemp)
	prometheus.MustRegister(owWeatherFeelsLike)
	prometheus.MustRegister(owWeatherTempMin)
	prometheus.MustRegister(owWeatherTempMax)
	prometheus.MustRegister(owWeatherPressure)
	prometheus.MustRegister(owWeatherHumidity)
	prometheus.MustRegister(owWeatherSeaLevel)
	prometheus.MustRegister(owWeatherGrndLevel)
	prometheus.MustRegister(owWeatherVisibility)
	prometheus.MustRegister(owWeatherWindSpeed)
	prometheus.MustRegister(owWeatherWindDeg)
	prometheus.MustRegister(owWeatherClouds)
	prometheus.MustRegister(owWeatherCondition)

	// Register air pollution metrics
	prometheus.MustRegister(owAirPollutionAQI)
	prometheus.MustRegister(owAirPollutionCO)
	prometheus.MustRegister(owAirPollutionNO)
	prometheus.MustRegister(owAirPollutionNO2)
	prometheus.MustRegister(owAirPollutionO3)
	prometheus.MustRegister(owAirPollutionSO2)
	prometheus.MustRegister(owAirPollutionPM25)
	prometheus.MustRegister(owAirPollutionPM10)
	prometheus.MustRegister(owAirPollutionNH3)
}

func fetchWeatherData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("weather API returned status code: %d", resp.StatusCode)
	}

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "", fmt.Errorf("failed to decode weather response: %w", err)
	}

	station := strconv.Itoa(weather.ID)

	// Update weather metrics
	owWeatherTemp.WithLabelValues(station).Set(weather.Main.Temp)
	owWeatherFeelsLike.WithLabelValues(station).Set(weather.Main.FeelsLike)
	owWeatherTempMin.WithLabelValues(station).Set(weather.Main.TempMin)
	owWeatherTempMax.WithLabelValues(station).Set(weather.Main.TempMax)
	owWeatherPressure.WithLabelValues(station).Set(weather.Main.Pressure)
	owWeatherHumidity.WithLabelValues(station).Set(weather.Main.Humidity)
	owWeatherSeaLevel.WithLabelValues(station).Set(weather.Main.SeaLevel)
	owWeatherGrndLevel.WithLabelValues(station).Set(weather.Main.GrndLevel)
	owWeatherVisibility.WithLabelValues(station).Set(weather.Visibility)
	owWeatherWindSpeed.WithLabelValues(station).Set(weather.Wind.Speed)
	owWeatherWindDeg.WithLabelValues(station).Set(weather.Wind.Deg)
	owWeatherClouds.WithLabelValues(station).Set(weather.Clouds.All)

	// Update weather condition (set to 1 to indicate active, 0 would be inactive)
	if len(weather.Weather) > 0 {
		owWeatherCondition.WithLabelValues(station, weather.Weather[0].Main, weather.Weather[0].Description).Set(1)
	}

	return station, nil
}

func fetchAirPollutionData(url string, station string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch air pollution data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("air pollution API returned status code: %d", resp.StatusCode)
	}

	var pollution AirPollutionResponse
	if err := json.NewDecoder(resp.Body).Decode(&pollution); err != nil {
		return fmt.Errorf("failed to decode air pollution response: %w", err)
	}

	// Update air pollution metrics
	if len(pollution.List) > 0 {
		data := pollution.List[0]
		owAirPollutionAQI.WithLabelValues(station).Set(float64(data.Main.AQI))
		owAirPollutionCO.WithLabelValues(station).Set(data.Components.CO)
		owAirPollutionNO.WithLabelValues(station).Set(data.Components.NO)
		owAirPollutionNO2.WithLabelValues(station).Set(data.Components.NO2)
		owAirPollutionO3.WithLabelValues(station).Set(data.Components.O3)
		owAirPollutionSO2.WithLabelValues(station).Set(data.Components.SO2)
		owAirPollutionPM25.WithLabelValues(station).Set(data.Components.PM25)
		owAirPollutionPM10.WithLabelValues(station).Set(data.Components.PM10)
		owAirPollutionNH3.WithLabelValues(station).Set(data.Components.NH3)
	}

	return nil
}

func updateMetrics(weatherURL, pollutionURL string) {
	station, err := fetchWeatherData(weatherURL)
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		return
	}

	if err := fetchAirPollutionData(pollutionURL, station); err != nil {
		log.Printf("Error fetching air pollution data: %v", err)
	}
}

func main() {
	// Load environment variables from .env if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables: %v", err)
	}

	latitude := os.Getenv("LATITUDE")
	longitude := os.Getenv("LONGITUDE")
	units := os.Getenv("UNITS")
	if units == "" {
		units = "standard"
	}
	if units != "standard" && units != "imperial" && units != "metric" {
		log.Fatal("UNITS must be either standard, imperial, or metric")
	}
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	exporterPort := os.Getenv("EXPORTER_PORT")

	if latitude == "" || longitude == "" || apiKey == "" {
		log.Fatal("LATITUDE, LONGITUDE, and OPENWEATHER_API_KEY environment variables must be set")
	}

	if exporterPort == "" {
		exporterPort = "8080"
	}

	currentWeatherURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%s&lon=%s&appid=%s&units=%s", latitude, longitude, apiKey, units)
	pollutionURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/air_pollution?lat=%s&lon=%s&appid=%s", latitude, longitude, apiKey)

	// Initial fetch
	updateMetrics(currentWeatherURL, pollutionURL)

	// Update metrics every 5 minutes
	go func() {
		// Update metrics every 5 minutes
		// 2 API calls per tick, 576 calls per day, below the 1000 limit for the free tier
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			updateMetrics(currentWeatherURL, pollutionURL)
		}
	}()

	// Set up HTTP server for metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>OpenWeather Exporter</title></head>
			<body>
				<h1>OpenWeather Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
			</body>
		</html>`))
	})

	log.Printf("Starting OpenWeather exporter on port %s", exporterPort)
	log.Fatal(http.ListenAndServe(":"+exporterPort, nil))
}
