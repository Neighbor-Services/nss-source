package main

import (
	"fmt"
	"log"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CategoryData struct {
	Name        string
	Description string
	Image       string
	Services    []ServiceData
}

type ServiceData struct {
	Name        string
	Description string
	BasePrice   float64
}

var categoriesSeed = []CategoryData{
	{
		Name:        "Beauty & Grooming",
		Description: "Mobile and location-based hair, barber, nail, makeup, and spa wellness services.",
		Image:       "categories/beauty_grooming.jpg",
		Services: []ServiceData{
			{"Mobile Barber", "On-demand haircut, fade, beard sculpting and hot towel shaves at your home or office.", 35.00},
			{"Mobile Hairstylist", "Professional blowout, cutting, styling, coloring and hair treatments on location.", 50.00},
			{"Mobile Nail Technician", "In-home manicure, pedicure, gel nails, acrylics and custom nail art.", 40.00},
			{"Mobile Makeup Artist", "Bridal, event, photoshoot and glam makeup application at your doorstep.", 60.00},
			{"Mobile Spa / Massage Therapist", "Traveling therapeutic massage, deep tissue, swedish and relaxing spa bodywork.", 80.00},
			{"Mobile Braider / Loctician", "Box braids, knotless braids, cornrows, dreadlocks and loc maintenance.", 75.00},
			{"Mobile Beautician", "Facials, skin glow treatments, lash extensions, brow shaping and waxing.", 45.00},
			{"Barbershop", "Traditional and modern barbershop walk-in and appointment hair cutting.", 30.00},
			{"Hair & Beauty Salon", "Full-service salon hair styling, extensions, coloring, and wash & set.", 45.00},
			{"Nail Salon", "Full set acrylics, dip powder, spa pedicure and manicure studio services.", 35.00},
			{"Waxing Studio", "Body waxing, brazilian, facial waxing and precision eyebrow shaping.", 40.00},
			{"Massage Parlor & Clinic", "Licensed massage clinic bodywork, pain relief and relaxation therapies.", 75.00},
		},
	},
	{
		Name:        "Healthcare & Wellness",
		Description: "Mobile medical, diagnostic labs, dental, physical therapy, hydration, and clinic care.",
		Image:       "categories/healthcare.jpg",
		Services: []ServiceData{
			{"Mobile Primary Care (NP/PA)", "Routine health assessments, wellness checks and chronic condition monitoring.", 120.00},
			{"Mobile Urgent Care", "Same-day treatment for non-emergency injuries, acute illnesses and minor care.", 150.00},
			{"Mobile Dental Services", "Dental cleanings, exams, fluoride treatments and on-site oral care vans.", 100.00},
			{"Mobile Physical & Occupational Therapy", "Rehabilitation therapy, injury recovery and mobility exercises at home.", 90.00},
			{"Mobile Phlebotomy & Diagnostics", "In-home blood draws, specimen collection, lab testing and diagnostic panels.", 65.00},
			{"Mobile IV Hydration Therapy", "Vitamin infusions, athletic recovery, immunity boosts and hangover relief drips.", 110.00},
			{"Mobile Immunization Clinic", "Flu shots, routine vaccines, travel immunizations and community wellness.", 50.00},
			{"Mobile Medical Imaging (X-Ray / Ultrasound)", "On-demand mobile digital X-rays and ultrasound diagnostic scans.", 180.00},
			{"Family & Urgent Care Clinic", "Outpatient healthcare clinic visits, prescriptions and medical consults.", 100.00},
			{"Dental & Optical Clinic", "Comprehensive dentistry, fillings, eye exams, glasses and contact lenses.", 85.00},
			{"Physical Therapy & Chiropractic Center", "Spine adjustments, musculoskeletal alignment, posture correction and PT.", 75.00},
			{"Mental Health Counseling Center", "Licensed individual, couples, family counseling and psychotherapy.", 90.00},
			{"Pharmacy & Prescription Delivery", "Prescription medication dispensing, compounding and home delivery.", 25.00},
		},
	},
	{
		Name:        "Caregiving & Family",
		Description: "Trusted child care, babysitting, adult homecare, senior companionship, and nanny services.",
		Image:       "categories/family_care.jpg",
		Services: []ServiceData{
			{"Baby Sitter & Nanny", "Experienced and background-checked in-home childcare and babysitting.", 25.00},
			{"Child Care & Preschool", "Licensed daycare facilities, early childhood education and preschool programs.", 35.00},
			{"Adult Homecare & Senior Companion", "Compassionate senior assistance, meal preparation, medication reminders and companion care.", 30.00},
			{"In-Home Caregiver", "Personal hygiene support, mobility assistance, recovery care and nursing aides.", 35.00},
		},
	},
	{
		Name:        "Pet Care & Grooming",
		Description: "Mobile dog grooming, pet sitting, dog walking, and behavioral training.",
		Image:       "categories/pet_care.jpg",
		Services: []ServiceData{
			{"Mobile Dog Grooming", "Hydrobath, breed-specific haircut, deshedding, ear cleaning and nail clipping in a mobile van.", 55.00},
			{"Mobile Pet Grooming", "Cat grooming, coat conditioning, teeth brushing and gentle pet sanitation.", 50.00},
			{"Dog Walker & Sitter", "Daily dog walks, pet drop-in visits, feeding, playtime and overnight sitting.", 25.00},
			{"Dog & Puppy Training", "Basic obedience, leash manners, puppy socialization and behavioral modification.", 65.00},
			{"Pet Boarding & Daycare", "Safe, cage-free pet hotel boarding, supervised play and overnight care.", 40.00},
		},
	},
	{
		Name:        "Automotive & Transport",
		Description: "Mobile mechanics, oil changes, tire repair, detailing, car buyers, and private chauffeurs.",
		Image:       "categories/automotive.jpg",
		Services: []ServiceData{
			{"Mobile Mechanic", "On-site engine diagnostics, starter/alternator replacement, tune-ups and brake jobs.", 75.00},
			{"Mobile Oil Change", "Synthetic and standard oil changes, oil filter replacement, and fluid top-off.", 45.00},
			{"Mobile Tire Service", "Flat tire repair, tire mounting, balancing, rotation and emergency roadside tire change.", 50.00},
			{"Mobile Windshield Repair", "Windshield chip and crack repair, auto glass replacement and seal inspection.", 60.00},
			{"Mobile Car Wash & Detailing", "Hand wash, interior steam cleaning, leather conditioning, wax and paint sealant.", 40.00},
			{"Auto Repair Shop", "Full mechanical auto garage repairs, transmission service, alignment and state inspections.", 80.00},
			{"Tire Shop & Alignment", "New and used tire sales, wheel balancing, alignment and seasonal tire changeover.", 45.00},
			{"Car Battery Replacement", "Battery testing, mobile delivery, terminal cleaning and new battery installation.", 35.00},
			{"Concierge Car Buyer", "Expert vehicle price negotiation, dealer liaising, and pre-purchase inspection assistance.", 150.00},
			{"Personal Chauffeur", "Professional private driver for events, business trips, airport runs and daily travel.", 45.00},
			{"Delivery Driver & Courier", "Fast package delivery, grocery transport, document dispatch and cargo hauling.", 20.00},
		},
	},
	{
		Name:        "Food & Culinary",
		Description: "Personal chefs, catering, food trucks, bakeries, restaurants, and cloud kitchens.",
		Image:       "categories/food_culinary.jpg",
		Services: []ServiceData{
			{"Personal Chef", "Custom in-home meal prep, multi-course private dinner parties and dietary cuisine.", 100.00},
			{"Event Catering Services", "Buffet and plated catering for weddings, corporate events, parties and celebrations.", 200.00},
			{"Mobile Food Truck", "Gourmet street food, burgers, tacos, BBQ and specialty dishes on wheels.", 150.00},
			{"Mobile Coffee Van", "Espresso bar on wheels, cold brew, lattes, pastries and event barista service.", 100.00},
			{"Mobile Ice Cream Truck", "Artisan ice cream, popsicles, soft serve and dessert catering for events.", 75.00},
			{"Restaurant & Café", "Dine-in, takeout, breakfast, brunch, lunch and fine dinner experiences.", 30.00},
			{"Bakery & Dessert Shop", "Custom cakes, artisan breads, cupcakes, cookies, pastries and dessert platters.", 25.00},
			{"Cloud & Catering Kitchen", "Commercial kitchen rentals, ghost kitchen food production and delivery hub.", 150.00},
			{"Grocery Store & Supermarket", "Fresh produce, pantry staples, household essentials and grocery delivery.", 15.00},
		},
	},
	{
		Name:        "Home Improvement & Repair",
		Description: "Handyman, home inspection, locksmith, roofing, painting, plumbing, and electrical.",
		Image:       "categories/home_repair.jpg",
		Services: []ServiceData{
			{"Mobile Handyman", "Drywall patching, door knob repair, light fixture hanging and general repairs.", 50.00},
			{"Home Inspection", "Pre-purchase home, roof, electrical, plumbing, foundation and radon inspections.", 250.00},
			{"Locksmith", "Emergency lockout service, deadbolt installation, smart lock setup and lock rekeying.", 65.00},
			{"TV Mounting & Shelf Installation", "Secure TV wall mounting, floating shelves, art hanging and wire concealment.", 60.00},
			{"Mobile Furniture Assembly & Repair", "IKEA and furniture assembly, desk, bed, wardrobe, table, chair assembly & woodwork.", 45.00},
			{"Roof Repair", "Shingle replacement, roof leak repairs, chimney flashing, soffit and siding repair.", 120.00},
			{"Gutter Cleaning & Installation", "Downspout clearing, seamless gutter installation and leaf guard protection.", 75.00},
			{"Interior & Exterior Wall Painting", "Residential painting, trim staining, accent walls, wallpapering and drywall primer.", 100.00},
			{"Plumbing & Leak Repair", "Pipe leak fix, faucet replacement, clogged drain clearing, toilets, and water heaters.", 85.00},
			{"Electrical & Wiring", "Ceiling fan install, breaker panel repair, outlet replacement, switches and lighting.", 90.00},
			{"HVAC & AC Repair", "Air conditioning repair, furnace maintenance, thermostat wiring and duct cleaning.", 95.00},
		},
	},
	{
		Name:        "Cleaning & Sanitation",
		Description: "House cleaning, deep clean, window washing, debris removal, and carpet cleaning.",
		Image:       "categories/cleaning.jpg",
		Services: []ServiceData{
			{"House Cleaning", "Standard recurring house cleaning, kitchen scrubbing, bathroom sanitizing and dusting.", 70.00},
			{"Deep House Clean", "Detailed top-to-bottom scrub, appliance interior cleaning, baseboards and move-out cleans.", 120.00},
			{"Window Washer", "Interior and exterior window pane washing, screen cleaning and track detailing.", 50.00},
			{"Carpet & Upholstery Cleaning", "Hot water extraction steam cleaning, stain removal, sofa and rug sanitizing.", 65.00},
			{"Debris Removal Service", "Post-construction cleanout, junk removal, yard debris haul and garage clearing.", 95.00},
			{"Commercial Janitorial & Sanitation", "Office building sanitization, commercial floor buffing, restrooms and trash service.", 110.00},
		},
	},
	{
		Name:        "Lawn & Pest Control",
		Description: "Lawn mowing, tree trimming, landscaping design, and pest extermination.",
		Image:       "categories/lawn_pest.jpg",
		Services: []ServiceData{
			{"Lawn Care & Mowing", "Weekly lawn mowing, edging, weed whacking, fertilization, seeding and aeration.", 40.00},
			{"Tree Service & Trimming", "Tree pruning, hazardous branch removal, stump grinding and tree felling.", 150.00},
			{"Landscaping & Garden Design", "Mulch delivery, flower bed planting, sod installation, patio pavers and irrigation.", 80.00},
			{"Pest Control & Extermination", "Termite treatments, rodent baiting, bed bug heat treatments, ant and roach sprays.", 85.00},
		},
	},
	{
		Name:        "Fashion & Apparel",
		Description: "Dressmakers, alterations, dry cleaning, fashion consultants, and personal shoppers.",
		Image:       "categories/fashion.jpg",
		Services: []ServiceData{
			{"Dress Maker & Tailoring", "Custom gown creation, suit bespoke tailoring, dressmaking and couture sewing.", 50.00},
			{"Clothing Alterations", "Pant hemming, zipper replacement, waist resizing, suit tapering and seam mending.", 25.00},
			{"Dry Cleaning & Laundry", "Professional dry cleaning, pressed shirts, garment steaming and wash & fold.", 20.00},
			{"Fashion Consultant", "Wardrobe audit, personal styling, color analysis and event outfit coordination.", 60.00},
			{"Personal Shopper", "In-store concierge shopping, gift procurement, luxury sourcing and grocery runs.", 35.00},
		},
	},
	{
		Name:        "Professional Services",
		Description: "Realtors, mobile notaries, accountants, tax preparation, and IT tech support.",
		Image:       "categories/professional.jpg",
		Services: []ServiceData{
			{"Realtor & Real Estate", "Buyer representation, seller listing, property valuation, showings and lease contracts.", 100.00},
			{"Mobile Notary", "Traveling loan signing agent, document notarization, affidavits and powers of attorney.", 35.00},
			{"Mobile Accounting & Tax Prep", "Individual & business tax returns, bookkeeping, payroll and CPA financial advisory.", 75.00},
			{"Mobile IT Support", "Wi-Fi network setup, computer virus removal, printer configuration and tech troubleshooting.", 60.00},
		},
	},
	{
		Name:        "Education & Digital",
		Description: "Tutoring, music lessons, fitness gyms, web design, digital marketing, and virtual assistants.",
		Image:       "categories/education_digital.jpg",
		Services: []ServiceData{
			{"Academic Tutoring & Test Prep", "K-12 math, science, english tutoring, SAT/ACT test prep and college readiness.", 40.00},
			{"Music & Arts Instruction", "Piano, guitar, voice, violin lessons, painting, drawing and pottery classes.", 45.00},
			{"Gym & Personal Fitness Training", "1-on-1 personal training, strength coaching, weight loss programs and nutrition plans.", 50.00},
			{"Yoga & Martial Arts Coaching", "Private yoga sessions, karate, jiu-jitsu, kickboxing and self-defense coaching.", 45.00},
			{"Digital Marketing & Social Media", "Social media management, SEO optimization, Facebook/Google ads and email marketing.", 150.00},
			{"Graphic Design & Branding", "Logo design, brand identity systems, marketing flyers, menus and business cards.", 100.00},
			{"Web & Mobile App Development", "Custom website development, Flutter/React mobile apps, UI/UX and maintenance.", 250.00},
			{"Virtual Assistant", "Email inbox management, appointment scheduling, data entry, research and admin tasks.", 25.00},
			{"Online Business Consulting", "Startup strategy, operations optimization, executive leadership and business plans.", 100.00},
		},
	},
}

func main() {
	log.Println("🚀 Starting Neighbor Service (NS) Categories & Catalog Services Seed Script...")

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	var totalCategories int
	var totalServices int

	for _, catData := range categoriesSeed {
		// 1. Find or create Category by unique Name
		var category entity.Category
		err := db.Where("name = ?", catData.Name).First(&category).Error
		if err == gorm.ErrRecordNotFound {
			category = entity.Category{
				ID:          uuid.New(),
				Name:        catData.Name,
				Description: catData.Description,
				Image:       catData.Image,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
			}
			if err := db.Create(&category).Error; err != nil {
				log.Fatalf("❌ Failed to create category '%s': %v", catData.Name, err)
			}
			log.Printf("✨ Created new Category: %s", category.Name)
		} else if err != nil {
			log.Fatalf("❌ Error querying category '%s': %v", catData.Name, err)
		} else {
			// Update description/image
			category.Description = catData.Description
			if category.Image == "" {
				category.Image = catData.Image
			}
			category.UpdatedAt = time.Now().UTC()
			db.Save(&category)
			log.Printf("🔄 Updated existing Category: %s", category.Name)
		}
		totalCategories++

		// 2. Insert or update Catalog Services
		for _, sData := range catData.Services {
			var catalogSvc entity.CatalogService
			err := db.Where("name = ? AND category_id = ?", sData.Name, category.ID).First(&catalogSvc).Error
			priceVal := sData.BasePrice

			if err == gorm.ErrRecordNotFound {
				catalogSvc = entity.CatalogService{
					ID:          uuid.New(),
					CategoryID:  category.ID,
					Name:        sData.Name,
					Description: sData.Description,
					BasePrice:   &priceVal,
					CreatedAt:   time.Now().UTC(),
					UpdatedAt:   time.Now().UTC(),
				}
				if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&catalogSvc).Error; err != nil {
					log.Printf("⚠️ Could not create service '%s': %v", sData.Name, err)
					continue
				}
				log.Printf("   ➕ Added Catalog Service: [%s] -> %s ($%.2f)", category.Name, catalogSvc.Name, priceVal)
			} else if err != nil {
				log.Printf("⚠️ Error querying service '%s': %v", sData.Name, err)
			} else {
				catalogSvc.Description = sData.Description
				if catalogSvc.BasePrice == nil {
					catalogSvc.BasePrice = &priceVal
				}
				catalogSvc.UpdatedAt = time.Now().UTC()
				db.Save(&catalogSvc)
			}
			totalServices++
		}
	}

	fmt.Println("==========================================================")
	log.Printf("🎉 Seed Complete! Processed %d Categories and %d Catalog Services successfully.\n", totalCategories, totalServices)
	fmt.Println("==========================================================")
}

