# OpenWeather Prometheus Exporter

A Prometheus exporter that queries the OpenWeather API for current weather and air pollution data, exposing metrics for monitoring and alerting.

## Features

- **Weather Metrics**: Exposes current weather conditions including temperature, pressure, humidity, wind speed, visibility, and cloud coverage
- **Air Pollution Metrics**: Provides air quality data including AQI, CO, NO, NO2, O3, SO2, PM2.5, PM10, and NH3 concentrations
- **Environment Variable Support**: Can read configuration from `.env` file or system environment variables
- **Docker Support**: Includes Dockerfile for containerized deployment
- **Units Selection**: Supports the API's standard units, metric, and imperial units. See the metrics table below for details.

## Prerequisites

- Go 1.25.1 or later
- OpenWeather API key ([Get one here](https://openweathermap.org/api))
- Prometheus (for scraping metrics)

## Installation

### From Source

1. Clone the repository:
```bash
git clone <repository-url>
cd openweather_exporter
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o openweather_exporter main.go
```

## Configuration

The exporter requires the following environment variables:

### Required Variables

- `LATITUDE`: Latitude coordinate of the location
- `LONGITUDE`: Longitude coordinate of the location
- `OPENWEATHER_API_KEY`: Your OpenWeather API key

### Optional Variables

- `UNITS`: Temperature unit system (`standard`, `metric`, or `imperial`)
  - `standard`: Temperature in Kelvin (default), otherwise same as metric
  - `metric`: Temperature in Celsius, all other units standard metric
  - `imperial`: Temperature in Fahrenheit, speed in miles/hour, all other units are metric
- `EXPORTER_PORT`: Port for the HTTP server (default: `8080`)

### Configuration via .env File

Create a `.env` file in the project root:

```env
LATITUDE=32.27
LONGITUDE=-112.73
OPENWEATHER_API_KEY=your_api_key_here
UNITS=metric
EXPORTER_PORT=8080
```

The exporter will automatically load variables from the `.env` file if it exists. If the file is not found, it will use system environment variables.

## Usage

### Running Locally

```bash
# Using .env file
./openweather_exporter

# Or with environment variables
export LATITUDE=32.27
export LONGITUDE=-112.73
export OPENWEATHER_API_KEY=your_api_key
export UNITS=metric
./openweather_exporter
```

### Running with Docker

1. Build the Docker image:
```bash
docker build -t openweather_exporter .
```

2. Run the container:
```bash
# Using environment variables
docker run -d \
  -p 8080:8080 \
  -e LATITUDE=32.27 \
  -e LONGITUDE=-112.73 \
  -e OPENWEATHER_API_KEY=your_api_key \
  -e UNITS=metric \
  openweather_exporter

# Or using .env file
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  openweather_exporter

# Or using the docker compose file
docker compose -f docker_compose.yaml up -d
```

Note: The Dockerfile supports multi-platform builds and will automatically build for your system's architecture. For multi-platform builds, use:
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t openweather_exporter .
```

## API Endpoints

- `GET /`: Simple HTML page with a link to metrics
- `GET /metrics`: Prometheus metrics endpoint

## Metrics

All metrics are labeled with `station` (the weather station ID from OpenWeather).

### Weather Metrics (prefix: `ow_weather_`)

| Metric | Description | Unit |
|--------|-------------|------|
| `ow_weather_temp` | Current temperature | Depends on UNITS setting |
| `ow_weather_feels_like` | Feels like temperature | Depends on UNITS setting |
| `ow_weather_temp_min` | Minimum temperature | Depends on UNITS setting |
| `ow_weather_temp_max` | Maximum temperature | Depends on UNITS setting |
| `ow_weather_pressure` | Atmospheric pressure | hPa |
| `ow_weather_humidity` | Humidity percentage | % |
| `ow_weather_sea_level` | Sea level pressure | hPa |
| `ow_weather_grnd_level` | Ground level pressure | hPa |
| `ow_weather_visibility` | Visibility | meters |
| `ow_weather_wind_speed` | Wind speed | Depends on UNITS setting |
| `ow_weather_wind_deg` | Wind direction | degrees |
| `ow_weather_clouds` | Cloud coverage | % |
| `ow_weather_condition` | Weather condition (1 = active) | - |

The `ow_weather_condition` metric includes additional labels:
- `main`: Main weather condition (e.g., "Clear", "Clouds", "Rain")
- `description`: Detailed description (e.g., "clear sky", "light rain")

### Air Pollution Metrics (prefix: `ow_air_pollution_`)

| Metric | Description | Unit |
|--------|-------------|------|
| `ow_air_pollution_aqi` | Air Quality Index | 1-5 |
| `ow_air_pollution_co` | Carbon monoxide | μg/m³ |
| `ow_air_pollution_no` | Nitrogen monoxide | μg/m³ |
| `ow_air_pollution_no2` | Nitrogen dioxide | μg/m³ |
| `ow_air_pollution_o3` | Ozone | μg/m³ |
| `ow_air_pollution_so2` | Sulphur dioxide | μg/m³ |
| `ow_air_pollution_pm2_5` | PM2.5 particles | μg/m³ |
| `ow_air_pollution_pm10` | PM10 particles | μg/m³ |
| `ow_air_pollution_nh3` | Ammonia | μg/m³ |

## Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'weather-exporter'
    static_configs:
      - targets: ['localhost:8080']
```

## API Rate Limits

The exporter makes 2 API calls every 5 minutes (one for weather, one for air pollution), resulting in:
- 24 calls per hour
- 576 calls per day

This is well below the free tier limit of 1,000 calls per day.
