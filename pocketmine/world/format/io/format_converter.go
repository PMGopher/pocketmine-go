package io

import (
	"crypto/rand"
	"fmt"
	"hash/crc32"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
)

// FormatConverter is a port of pocketmine\world\format\io\FormatConverter: converts a world to
// another (writable) format, backing up the original.
type FormatConverter struct {
	oldProvider             WorldProvider
	newProvider             *WritableWorldProviderManagerEntry
	backupPath              string
	logger                  log.Logger
	chunksPerProgressUpdate int
	// generatorName is GeneratorManager::getInstance()->getGenerator($name)?->getGeneratorClass()
	// followed by getGeneratorName(): the registered name of a generator alias, or false if
	// unknown (the world package owns the generator registry).
	generatorName func(name string) (string, bool)
}

// NewFormatConverter is a port of FormatConverter::__construct (chunksPerProgressUpdate defaults
// to 256 in PHP).
func NewFormatConverter(oldProvider WorldProvider, newProvider *WritableWorldProviderManagerEntry, backupPath string, logger log.Logger, chunksPerProgressUpdate int, generatorName func(name string) (string, bool)) *FormatConverter {
	c := &FormatConverter{
		oldProvider:             oldProvider,
		newProvider:             newProvider,
		logger:                  log.NewPrefixedLogger(logger, "World Converter: "+oldProvider.GetWorldData().GetName()),
		chunksPerProgressUpdate: chunksPerProgressUpdate,
		generatorName:           generatorName,
	}
	if _, err := os.Stat(backupPath); err != nil {
		_ = os.MkdirAll(backupPath, 0o777)
	}
	nextSuffix := ""
	for {
		c.backupPath = filepath.Join(backupPath, filepath.Base(oldProvider.GetPath())+nextSuffix)
		random := make([]byte, 4)
		_, _ = rand.Read(random)
		nextSuffix = fmt.Sprintf("_%d", crc32.ChecksumIEEE(random))
		if _, err := os.Stat(c.backupPath); err != nil {
			break
		}
	}
	return c
}

// GetBackupPath is a port of FormatConverter::getBackupPath.
func (c *FormatConverter) GetBackupPath() string { return c.backupPath }

// Execute is a port of FormatConverter::execute.
func (c *FormatConverter) Execute() error {
	newProvider, err := c.generateNew()
	if err != nil {
		return err
	}

	if err := c.populateLevelData(newProvider.GetWorldData()); err != nil {
		return err
	}
	if err := c.convertTerrain(newProvider); err != nil {
		return err
	}

	path := c.oldProvider.GetPath()
	_ = c.oldProvider.Close()
	_ = newProvider.Close()

	c.logger.Info("Backing up pre-conversion world to " + c.backupPath)
	if err := os.Rename(path, c.backupPath); err != nil {
		c.logger.Warning("Moving old world files for backup failed, attempting copy instead. This might take a long time.")
		if err := utils.RecursiveCopy(path, c.backupPath); err != nil {
			return err
		}
		if err := utils.RecursiveUnlink(path); err != nil {
			return err
		}
	}
	if err := os.Rename(newProvider.GetPath(), path); err != nil {
		// we don't expect this to happen because worlds/ should most likely be all on the same FS, but just in case...
		c.logger.Debug("Relocation of new world files to location failed, attempting copy and delete instead")
		if err := utils.RecursiveCopy(newProvider.GetPath(), path); err != nil {
			return err
		}
		if err := utils.RecursiveUnlink(newProvider.GetPath()); err != nil {
			return err
		}
	}

	c.logger.Info("Conversion completed")
	return nil
}

func (c *FormatConverter) generateNew() (WritableWorldProvider, error) {
	c.logger.Info("Generating new world")
	data := c.oldProvider.GetWorldData()

	convertedOutput := strings.TrimRight(c.oldProvider.GetPath(), "/"+string(os.PathSeparator)) + "_converted" + string(os.PathSeparator)
	if _, err := os.Stat(convertedOutput); err == nil {
		c.logger.Info("Found previous conversion attempt, deleting...")
		if err := utils.RecursiveUnlink(convertedOutput); err != nil {
			return nil, err
		}
	}
	//TODO: defaulting to NORMAL here really isn't very good behaviour, but it's consistent with what we already
	//did previously; besides, WorldManager checks for unknown generators before this is reached anyway.
	generatorName := "normal"
	if c.generatorName != nil {
		if name, ok := c.generatorName(data.GetGenerator()); ok {
			generatorName = name
		}
	}
	if err := os.MkdirAll(convertedOutput, 0o777); err != nil {
		return nil, err
	}
	if err := c.newProvider.Generate(convertedOutput, data.GetName(), WorldCreationOptions{
		GeneratorName:    generatorName,
		GeneratorOptions: data.GetGeneratorOptions(),
		Seed:             data.GetSeed(),
		SpawnPosition:    data.GetSpawn(),
		Difficulty:       data.GetDifficulty(),
	}); err != nil {
		return nil, err
	}
	return c.newProvider.FromPathWritable(convertedOutput, c.logger)
}

func (c *FormatConverter) populateLevelData(data WorldData) error {
	c.logger.Info("Converting world manifest")
	oldData := c.oldProvider.GetWorldData()
	data.SetDifficulty(oldData.GetDifficulty())
	data.SetLightningLevel(oldData.GetLightningLevel())
	data.SetLightningTime(oldData.GetLightningTime())
	data.SetRainLevel(oldData.GetRainLevel())
	data.SetRainTime(oldData.GetRainTime())
	data.SetSpawn(oldData.GetSpawn())
	data.SetTime(oldData.GetTime())

	if err := data.Save(); err != nil {
		return err
	}
	c.logger.Info("Finished converting manifest")
	//TODO: add more properties as-needed
	return nil
}

func (c *FormatConverter) convertTerrain(newProvider WritableWorldProvider) error {
	c.logger.Info("Calculating chunk count")
	count, err := c.oldProvider.CalculateChunkCount()
	if err != nil {
		return err
	}
	c.logger.Info(fmt.Sprintf("Discovered %d chunks", count))

	counter := 0
	start := time.Now()
	thisRound := start
	var saveErr error
	err = c.oldProvider.GetAllChunks(true, c.logger, func(coords ChunkCoords, loadedChunkData *LoadedChunkData) bool {
		if saveErr = newProvider.SaveChunk(coords.X, coords.Z, loadedChunkData.GetData(), format.DirtyFlagsAll); saveErr != nil {
			return false
		}
		counter++
		if counter%c.chunksPerProgressUpdate == 0 {
			now := time.Now()
			diff := now.Sub(thisRound).Seconds()
			thisRound = now
			c.logger.Info(fmt.Sprintf("Converted %d / %d chunks (%.0f chunks/sec)", counter, count, math.Floor(float64(c.chunksPerProgressUpdate)/diff)))
		}
		if counter%(1<<16) == 0 {
			newProvider.DoGarbageCollection()
		}
		return true
	})
	if err != nil {
		return err
	}
	if saveErr != nil {
		return saveErr
	}
	total := time.Since(start).Seconds()
	c.logger.Info(fmt.Sprintf("Converted %d / %d chunks in %.3f seconds (%.0f chunks/sec)", counter, counter, total, math.Floor(float64(counter)/total)))
	return nil
}
