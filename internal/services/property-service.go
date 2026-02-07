package services

import (
	"fmt"

	"github.com/adampresley/aletics/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PropertyServiceConfig struct {
	DB *gorm.DB
}

type PropertyService struct {
	db *gorm.DB
}

func NewPropertyService(config PropertyServiceConfig) *PropertyService {
	return &PropertyService{
		db: config.DB,
	}
}

func (s *PropertyService) ListProperties(filter string) ([]models.Property, error) {
	var (
		err        error
		properties []models.Property
		query      *gorm.DB
	)

	query = s.db.
		Model(&models.Property{}).
		Order("LOWER(name) asc")

	if filter != "" {
		query = query.Where("name LIKE ?", "%"+filter+"%")
	}

	if err = query.Find(&properties).Error; err != nil {
		return []models.Property{}, err
	}

	return properties, nil
}

func (s *PropertyService) GetProperty(id uint) (models.Property, error) {
	var (
		err      error
		property models.Property
	)

	if err = s.db.First(&property, id).Error; err != nil {
		return models.Property{}, err
	}

	return property, nil
}

func (s *PropertyService) CreateProperty(name, domain string) (models.Property, error) {
	var (
		err      error
		property models.Property
	)

	property = models.Property{
		Name:   name,
		Domain: domain,
		Token:  uuid.New().String(),
		Active: true,
	}

	if err = s.db.Create(&property).Error; err != nil {
		return models.Property{}, err
	}

	return property, nil
}

func (s *PropertyService) UpdateProperty(id uint, property models.Property) error {
	var (
		err              error
		existingProperty models.Property
	)

	if err = s.db.First(&existingProperty, id).Error; err != nil {
		return err
	}

	existingProperty.Name = property.Name
	existingProperty.Domain = property.Domain
	existingProperty.Active = property.Active

	fmt.Printf("\nexistingProperty: %+v\n", existingProperty)
	if err = s.db.Save(&existingProperty).Error; err != nil {
		return err
	}

	return nil
}

func (s *PropertyService) DeleteProperty(id uint) error {
	var (
		err error
	)

	if err = s.db.Delete(&models.Property{}, id).Error; err != nil {
		return err
	}

	return nil
}
