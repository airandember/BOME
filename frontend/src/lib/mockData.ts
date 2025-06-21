// COMPREHENSIVE MOCK DATA FOR BOME FRONTEND
// This file contains all mock data used throughout the application
// Replace with real API calls for production

import type { Video, VideoCategory, VideoComment } from './video';

// MOCK VIDEOS WITH BUNNY.NET URLs (simulating Bunny.net video streaming)
export const MOCK_VIDEOS: Video[] = [
	{
		id: 1,
		title: "Archaeological Evidence for the Book of Mormon",
		description: "Exploring recent archaeological discoveries that support Book of Mormon narratives, including ancient civilizations, metallurgy, and cultural practices found in the Americas.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/archaeological-evidence-bom/playlist.m3u8",
		duration: 942, // 15:42
		viewCount: 24567,
		likeCount: 1234,
		category: "Archaeology",
		tags: ["archaeology", "evidence", "ancient-america", "civilizations"],
		createdAt: "2024-01-15T10:30:00Z",
		updatedAt: "2024-01-15T10:30:00Z"
	},
	{
		id: 2,
		title: "DNA and the Book of Mormon: Scientific Perspectives",
		description: "A comprehensive look at DNA evidence and its relationship to Book of Mormon populations, examining recent genetic studies and their implications.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/dna-book-mormon-science/playlist.m3u8",
		duration: 1335, // 22:15
		viewCount: 18934,
		likeCount: 892,
		category: "Science",
		tags: ["dna", "science", "genetics", "populations"],
		createdAt: "2024-01-18T14:20:00Z",
		updatedAt: "2024-01-18T14:20:00Z"
	},
	{
		id: 3,
		title: "Mesoamerican Connections to Book of Mormon Geography",
		description: "Examining cultural and geographical connections between Mesoamerica and the Book of Mormon, including recent discoveries and scholarly research.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/mesoamerican-connections/playlist.m3u8",
		duration: 1113, // 18:33
		viewCount: 31245,
		likeCount: 1567,
		category: "Geography",
		tags: ["mesoamerica", "geography", "culture", "maya"],
		createdAt: "2024-01-20T09:45:00Z",
		updatedAt: "2024-01-20T09:45:00Z"
	},
	{
		id: 4,
		title: "Linguistic Analysis of Book of Mormon Names",
		description: "Scholarly analysis of Hebrew and Egyptian linguistic patterns in Book of Mormon names and their ancient Near Eastern connections.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/linguistic-analysis-names/playlist.m3u8",
		duration: 1518, // 25:18
		viewCount: 15678,
		likeCount: 743,
		category: "Linguistics",
		tags: ["linguistics", "hebrew", "names", "ancient-languages"],
		createdAt: "2024-01-22T16:12:00Z",
		updatedAt: "2024-01-22T16:12:00Z"
	},
	{
		id: 5,
		title: "Metallurgy in Ancient America: Book of Mormon Evidence",
		description: "Evidence of advanced metallurgy in pre-Columbian America and its relationship to Book of Mormon descriptions of metalworking.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/ancient-metallurgy/playlist.m3u8",
		duration: 1267, // 21:07
		viewCount: 12456,
		likeCount: 623,
		category: "Archaeology",
		tags: ["metallurgy", "ancient-technology", "archaeology", "metals"],
		createdAt: "2024-01-25T11:30:00Z",
		updatedAt: "2024-01-25T11:30:00Z"
	},
	{
		id: 6,
		title: "Ancient American Writing Systems and the Book of Mormon",
		description: "Exploring ancient writing systems found in the Americas and their potential connections to Book of Mormon record-keeping practices.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/ancient-writing-systems/playlist.m3u8",
		duration: 1789, // 29:49
		viewCount: 9876,
		likeCount: 456,
		category: "Linguistics",
		tags: ["writing-systems", "ancient-scripts", "record-keeping", "linguistics"],
		createdAt: "2024-01-28T13:45:00Z",
		updatedAt: "2024-01-28T13:45:00Z"
	},
	{
		id: 7,
		title: "Climate and Geography: Book of Mormon Lands",
		description: "Analysis of climate patterns and geographical features described in the Book of Mormon and their correlation with Mesoamerican regions.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/climate-geography/playlist.m3u8",
		duration: 1456, // 24:16
		viewCount: 21345,
		likeCount: 1089,
		category: "Geography",
		tags: ["climate", "geography", "mesoamerica", "environment"],
		createdAt: "2024-02-01T08:20:00Z",
		updatedAt: "2024-02-01T08:20:00Z"
	},
	{
		id: 8,
		title: "Warfare and Military Tactics in the Book of Mormon",
		description: "Examination of military strategies, fortifications, and warfare described in the Book of Mormon and their historical context.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/warfare-military-tactics/playlist.m3u8",
		duration: 1634, // 27:14
		viewCount: 17890,
		likeCount: 876,
		category: "History",
		tags: ["warfare", "military", "fortifications", "strategy"],
		createdAt: "2024-02-05T15:10:00Z",
		updatedAt: "2024-02-05T15:10:00Z"
	},
	{
		id: 9,
		title: "Agricultural Practices in Ancient America",
		description: "Evidence of sophisticated agricultural techniques in pre-Columbian America that align with Book of Mormon descriptions.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/agricultural-practices/playlist.m3u8",
		duration: 1123, // 18:43
		viewCount: 13567,
		likeCount: 678,
		category: "Agriculture",
		tags: ["agriculture", "farming", "ancient-techniques", "crops"],
		createdAt: "2024-02-08T12:00:00Z",
		updatedAt: "2024-02-08T12:00:00Z"
	},
	{
		id: 10,
		title: "Trade Networks in Ancient Mesoamerica",
		description: "Exploring extensive trade networks in ancient Mesoamerica and their relationship to Book of Mormon descriptions of commerce.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/trade-networks/playlist.m3u8",
		duration: 1398, // 23:18
		viewCount: 19234,
		likeCount: 934,
		category: "Economics",
		tags: ["trade", "commerce", "economy", "networks"],
		createdAt: "2024-02-12T14:30:00Z",
		updatedAt: "2024-02-12T14:30:00Z"
	},
	{
		id: 11,
		title: "Religious Practices and Temples in Ancient America",
		description: "Archaeological evidence of temple worship and religious practices in ancient America that parallel Book of Mormon descriptions.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/religious-practices-temples/playlist.m3u8",
		duration: 1876, // 31:16
		viewCount: 25678,
		likeCount: 1345,
		category: "Religion",
		tags: ["temples", "religion", "worship", "rituals"],
		createdAt: "2024-02-15T10:45:00Z",
		updatedAt: "2024-02-15T10:45:00Z"
	},
	{
		id: 12,
		title: "Cement and Construction in Ancient America",
		description: "Evidence of advanced construction techniques and cement use in pre-Columbian America, as mentioned in the Book of Mormon.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/cement-construction/playlist.m3u8",
		duration: 1245, // 20:45
		viewCount: 11234,
		likeCount: 567,
		category: "Archaeology",
		tags: ["construction", "cement", "building", "technology"],
		createdAt: "2024-02-18T16:20:00Z",
		updatedAt: "2024-02-18T16:20:00Z"
	},
	{
		id: 13,
		title: "Population Genetics and Migration Patterns",
		description: "Recent genetic studies revealing complex migration patterns to the Americas and their implications for Book of Mormon populations.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/population-genetics/playlist.m3u8",
		duration: 1567, // 26:07
		viewCount: 16789,
		likeCount: 823,
		category: "Science",
		tags: ["genetics", "migration", "populations", "dna"],
		createdAt: "2024-02-22T09:15:00Z",
		updatedAt: "2024-02-22T09:15:00Z"
	},
	{
		id: 14,
		title: "Ancient Roads and Transportation Systems",
		description: "Evidence of sophisticated road networks and transportation systems in ancient Mesoamerica mentioned in the Book of Mormon.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/ancient-roads-transportation/playlist.m3u8",
		duration: 1334, // 22:14
		viewCount: 14567,
		likeCount: 712,
		category: "Geography",
		tags: ["roads", "transportation", "infrastructure", "ancient-engineering"],
		createdAt: "2024-02-25T13:40:00Z",
		updatedAt: "2024-02-25T13:40:00Z"
	},
	{
		id: 15,
		title: "Textiles and Clothing in Ancient America",
		description: "Archaeological evidence of sophisticated textile production and clothing styles in ancient America as described in the Book of Mormon.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/textiles-clothing/playlist.m3u8",
		duration: 1189, // 19:49
		viewCount: 8934,
		likeCount: 445,
		category: "Culture",
		tags: ["textiles", "clothing", "culture", "crafts"],
		createdAt: "2024-02-28T11:25:00Z",
		updatedAt: "2024-02-28T11:25:00Z"
	},
	{
		id: 16,
		title: "Calendars and Astronomy in Ancient Mesoamerica",
		description: "Sophisticated astronomical knowledge and calendar systems in ancient Mesoamerica and their potential Book of Mormon connections.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/calendars-astronomy/playlist.m3u8",
		duration: 1723, // 28:43
		viewCount: 22456,
		likeCount: 1134,
		category: "Science",
		tags: ["astronomy", "calendars", "mathematics", "science"],
		createdAt: "2024-03-03T15:50:00Z",
		updatedAt: "2024-03-03T15:50:00Z"
	},
	{
		id: 17,
		title: "Disease and Health in Ancient American Populations",
		description: "Analysis of health patterns and disease in ancient American populations and their relationship to Book of Mormon accounts.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/disease-health-populations/playlist.m3u8",
		duration: 1456, // 24:16
		viewCount: 12789,
		likeCount: 634,
		category: "Science",
		tags: ["health", "disease", "populations", "medicine"],
		createdAt: "2024-03-06T12:30:00Z",
		updatedAt: "2024-03-06T12:30:00Z"
	},
	{
		id: 18,
		title: "Horses and Animals in Book of Mormon Lands",
		description: "Examining evidence for various animals mentioned in the Book of Mormon and their potential identifications in ancient America.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/horses-animals-bom/playlist.m3u8",
		duration: 1612, // 26:52
		viewCount: 18345,
		likeCount: 912,
		category: "Zoology",
		tags: ["animals", "horses", "fauna", "ecology"],
		createdAt: "2024-03-10T14:15:00Z",
		updatedAt: "2024-03-10T14:15:00Z"
	},
	{
		id: 19,
		title: "Government and Political Systems in the Book of Mormon",
		description: "Analysis of governmental structures and political systems described in the Book of Mormon and their historical parallels.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/government-political-systems/playlist.m3u8",
		duration: 1789, // 29:49
		viewCount: 15234,
		likeCount: 756,
		category: "History",
		tags: ["government", "politics", "leadership", "society"],
		createdAt: "2024-03-13T10:00:00Z",
		updatedAt: "2024-03-13T10:00:00Z"
	},
	{
		id: 20,
		title: "Maritime Technology and Ocean Travel",
		description: "Evidence of advanced maritime technology in ancient times and its relationship to Book of Mormon accounts of ocean voyages.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/maritime-technology/playlist.m3u8",
		duration: 1445, // 24:05
		viewCount: 13678,
		likeCount: 689,
		category: "Technology",
		tags: ["maritime", "ships", "ocean-travel", "navigation"],
		createdAt: "2024-03-16T16:45:00Z",
		updatedAt: "2024-03-16T16:45:00Z"
	},
	{
		id: 21,
		title: "Book of Mormon Witnesses: Historical Analysis",
		description: "Comprehensive examination of the testimonies and lives of the Book of Mormon witnesses and their historical credibility.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/bom-witnesses-analysis/playlist.m3u8",
		duration: 2145, // 35:45
		viewCount: 28934,
		likeCount: 1456,
		category: "History",
		tags: ["witnesses", "testimony", "historical", "credibility"],
		createdAt: "2024-03-20T11:30:00Z",
		updatedAt: "2024-03-20T11:30:00Z"
	},
	{
		id: 22,
		title: "Chiasmus and Literary Patterns in the Book of Mormon",
		description: "Analysis of complex literary structures, including chiasmus and other Hebrew poetic forms found in the Book of Mormon.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/chiasmus-literary-patterns/playlist.m3u8",
		duration: 1834, // 30:34
		viewCount: 19876,
		likeCount: 987,
		category: "Linguistics",
		tags: ["chiasmus", "literary", "hebrew", "poetry"],
		createdAt: "2024-03-23T14:15:00Z",
		updatedAt: "2024-03-23T14:15:00Z"
	},
	{
		id: 23,
		title: "Volcanic Activity and Book of Mormon Destruction",
		description: "Geological evidence of volcanic activity in Mesoamerica correlating with destruction described in 3 Nephi.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/volcanic-activity-destruction/playlist.m3u8",
		duration: 1523, // 25:23
		viewCount: 16543,
		likeCount: 821,
		category: "Geology",
		tags: ["volcanoes", "geology", "destruction", "3-nephi"],
		createdAt: "2024-03-26T09:45:00Z",
		updatedAt: "2024-03-26T09:45:00Z"
	},
	{
		id: 24,
		title: "Ancient American Coinage and Currency Systems",
		description: "Evidence of sophisticated monetary systems in ancient America and their relationship to Book of Mormon descriptions.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/ancient-coinage-currency/playlist.m3u8",
		duration: 1367, // 22:47
		viewCount: 12345,
		likeCount: 612,
		category: "Economics",
		tags: ["currency", "coinage", "economics", "monetary-systems"],
		createdAt: "2024-03-29T16:20:00Z",
		updatedAt: "2024-03-29T16:20:00Z"
	},
	{
		id: 25,
		title: "The Translation Process: Historical and Spiritual Perspectives",
		description: "Examination of the Book of Mormon translation process, historical accounts, and spiritual aspects of revelation.",
		thumbnailUrl: "/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png",
		videoUrl: "https://vz-12345678-123.b-cdn.net/videos/translation-process/playlist.m3u8",
		duration: 2234, // 37:14
		viewCount: 31567,
		likeCount: 1678,
		category: "Religion",
		tags: ["translation", "revelation", "spiritual", "historical"],
		createdAt: "2024-04-01T12:00:00Z",
		updatedAt: "2024-04-01T12:00:00Z"
	}
];

export const MOCK_CATEGORIES: VideoCategory[] = [
	{
		id: 1,
		name: "Archaeology",
		description: "Archaeological evidence and discoveries supporting Book of Mormon narratives",
		videoCount: 4
	},
	{
		id: 2,
		name: "Science",
		description: "Scientific perspectives and research related to Book of Mormon claims",
		videoCount: 5
	},
	{
		id: 3,
		name: "Geography",
		description: "Geographical studies and location theories for Book of Mormon lands",
		videoCount: 3
	},
	{
		id: 4,
		name: "Linguistics",
		description: "Language analysis and linguistic evidence from the Book of Mormon",
		videoCount: 3
	},
	{
		id: 5,
		name: "History",
		description: "Historical context and parallels to Book of Mormon events",
		videoCount: 3
	},
	{
		id: 6,
		name: "Culture",
		description: "Cultural practices and social structures in ancient America",
		videoCount: 1
	},
	{
		id: 7,
		name: "Agriculture",
		description: "Agricultural techniques and practices in ancient American civilizations",
		videoCount: 1
	},
	{
		id: 8,
		name: "Economics",
		description: "Economic systems and trade networks in ancient America",
		videoCount: 2
	},
	{
		id: 9,
		name: "Religion",
		description: "Religious practices and beliefs in ancient American cultures",
		videoCount: 2
	},
	{
		id: 10,
		name: "Technology",
		description: "Technological achievements and innovations in ancient America",
		videoCount: 1
	},
	{
		id: 11,
		name: "Zoology",
		description: "Animal life and fauna mentioned in the Book of Mormon",
		videoCount: 1
	},
	{
		id: 12,
		name: "Geology",
		description: "Geological evidence and natural phenomena in Book of Mormon accounts",
		videoCount: 1
	}
];

export const MOCK_COMMENTS: VideoComment[] = [
	{
		id: 1,
		videoId: 1,
		userId: 101,
		userName: "Dr. Sarah Mitchell",
		content: "Fascinating analysis! The archaeological evidence continues to support Book of Mormon narratives in remarkable ways.",
		createdAt: "2024-03-15T14:30:00Z"
	},
	{
		id: 2,
		videoId: 1,
		userId: 102,
		userName: "Michael Thompson",
		content: "This presentation really opened my eyes to the depth of evidence available. Thank you for the thorough research!",
		createdAt: "2024-03-15T15:45:00Z"
	},
	{
		id: 3,
		videoId: 1,
		userId: 103,
		userName: "Rebecca Johnson",
		content: "I appreciate the scholarly approach to this topic. The connections to Mesoamerican archaeology are compelling.",
		createdAt: "2024-03-15T16:20:00Z"
	},
	{
		id: 4,
		videoId: 2,
		userId: 104,
		userName: "Dr. James Wilson",
		content: "The genetic evidence is complex but fascinating. Great job explaining the nuances of DNA research.",
		createdAt: "2024-03-16T09:15:00Z"
	},
	{
		id: 5,
		videoId: 2,
		userId: 105,
		userName: "Lisa Chen",
		content: "This helps clarify many misconceptions about DNA and the Book of Mormon. Very informative!",
		createdAt: "2024-03-16T10:30:00Z"
	},
	{
		id: 6,
		videoId: 3,
		userId: 106,
		userName: "David Rodriguez",
		content: "The geographical correlations are amazing. I never realized how well Mesoamerica fits the Book of Mormon descriptions.",
		createdAt: "2024-03-17T13:45:00Z"
	},
	{
		id: 7,
		videoId: 3,
		userId: 107,
		userName: "Emily Davis",
		content: "Excellent research on the geographical aspects. The maps really help visualize the connections.",
		createdAt: "2024-03-17T14:20:00Z"
	},
	{
		id: 8,
		videoId: 4,
		userId: 108,
		userName: "Professor Mark Anderson",
		content: "The linguistic analysis is top-notch. The Hebrew and Egyptian connections are particularly intriguing.",
		createdAt: "2024-03-18T11:00:00Z"
	},
	{
		id: 9,
		videoId: 4,
		userId: 109,
		userName: "Jennifer Lee",
		content: "As someone studying ancient languages, I find this analysis very compelling and well-researched.",
		createdAt: "2024-03-18T12:15:00Z"
	},
	{
		id: 10,
		videoId: 5,
		userId: 110,
		userName: "Robert Taylor",
		content: "The metallurgy evidence is fascinating! I had no idea about the advanced metalworking in ancient America.",
		createdAt: "2024-03-19T16:30:00Z"
	},
	{
		id: 11,
		videoId: 6,
		userId: 111,
		userName: "Dr. Amanda Foster",
		content: "The writing systems research is excellent. The parallels to ancient Near Eastern scripts are remarkable.",
		createdAt: "2024-03-20T10:15:00Z"
	},
	{
		id: 12,
		videoId: 7,
		userId: 112,
		userName: "Carlos Mendez",
		content: "Great correlation between climate data and Book of Mormon descriptions. Very convincing!",
		createdAt: "2024-03-21T14:45:00Z"
	},
	{
		id: 13,
		videoId: 8,
		userId: 113,
		userName: "Captain John Harris",
		content: "As a military historian, I'm impressed by the accuracy of warfare descriptions in the Book of Mormon.",
		createdAt: "2024-03-22T11:30:00Z"
	},
	{
		id: 14,
		videoId: 9,
		userId: 114,
		userName: "Maria Gonzalez",
		content: "The agricultural evidence is compelling. Ancient American farming was much more sophisticated than I realized.",
		createdAt: "2024-03-23T09:20:00Z"
	},
	{
		id: 15,
		videoId: 10,
		userId: 115,
		userName: "Dr. William Chang",
		content: "Excellent analysis of trade networks. The economic complexity of ancient Mesoamerica is astounding.",
		createdAt: "2024-03-24T16:10:00Z"
	}
];

// MOCK USER DATA
export const MOCK_USERS = [
	{
		id: 1,
		email: "admin@bome.com",
		firstName: "Admin",
		lastName: "User",
		role: "admin",
		emailVerified: true,
		createdAt: "2024-01-01T00:00:00Z",
		lastLogin: "2024-04-01T12:00:00Z",
		status: "active",
		subscriptionStatus: "premium"
	},
	{
		id: 2,
		email: "john.doe@example.com",
		firstName: "John",
		lastName: "Doe",
		role: "user",
		emailVerified: true,
		createdAt: "2024-01-15T10:30:00Z",
		lastLogin: "2024-03-30T14:22:00Z",
		status: "active",
		subscriptionStatus: "basic"
	},
	{
		id: 3,
		email: "jane.smith@example.com",
		firstName: "Jane",
		lastName: "Smith",
		role: "user",
		emailVerified: true,
		createdAt: "2024-01-20T16:45:00Z",
		lastLogin: "2024-03-29T08:15:00Z",
		status: "active",
		subscriptionStatus: "premium"
	}
];

// MOCK DASHBOARD DATA
export const MOCK_DASHBOARD_DATA = {
	stats: {
		totalWatchTime: 1247,
		videosWatched: 23,
		favoriteVideos: 8,
		completedSeries: 3
	},
	recentActivity: [
		{
			type: "video_watched",
			title: "Archaeological Evidence for the Book of Mormon",
			timestamp: "2024-03-30T14:30:00Z"
		},
		{
			type: "video_liked",
			title: "DNA and the Book of Mormon",
			timestamp: "2024-03-29T16:45:00Z"
		},
		{
			type: "video_favorited",
			title: "Mesoamerican Connections",
			timestamp: "2024-03-28T11:20:00Z"
		}
	],
	recommendedVideos: MOCK_VIDEOS.slice(0, 6),
	favoriteVideos: MOCK_VIDEOS.slice(0, 4),
	continueWatching: [
		{
			...MOCK_VIDEOS[0],
			progress: 0.65,
			lastWatched: "2024-03-30T14:30:00Z"
		},
		{
			...MOCK_VIDEOS[2],
			progress: 0.23,
			lastWatched: "2024-03-29T20:15:00Z"
		}
	]
};

// MOCK API RESPONSE HELPERS
export const createMockResponse = (data: any, delay = 500) => {
	return new Promise((resolve) => {
		setTimeout(() => {
			resolve(data);
		}, delay);
	});
};

export const getMockVideos = (page = 1, limit = 20, category?: string, search?: string) => {
	let filteredVideos = [...MOCK_VIDEOS];

	// Filter by category
	if (category) {
		filteredVideos = filteredVideos.filter(video => 
			video.category.toLowerCase() === category.toLowerCase()
		);
	}

	// Filter by search query
	if (search) {
		const searchLower = search.toLowerCase();
		filteredVideos = filteredVideos.filter(video =>
			video.title.toLowerCase().includes(searchLower) ||
			video.description.toLowerCase().includes(searchLower) ||
			video.tags.some(tag => tag.toLowerCase().includes(searchLower))
		);
	}

	// Pagination
	const startIndex = (page - 1) * limit;
	const endIndex = startIndex + limit;
	const paginatedVideos = filteredVideos.slice(startIndex, endIndex);

	return {
		videos: paginatedVideos,
		pagination: {
			page,
			limit,
			total: filteredVideos.length,
			totalPages: Math.ceil(filteredVideos.length / limit)
		}
	};
};

export const getMockVideo = (id: number) => {
	const video = MOCK_VIDEOS.find(v => v.id === id);
	if (!video) {
		throw new Error('Video not found');
	}
	return { video };
};

export const getMockComments = (videoId: number, page = 1, limit = 20) => {
	const videoComments = MOCK_COMMENTS.filter(comment => comment.videoId === videoId);
	const startIndex = (page - 1) * limit;
	const endIndex = startIndex + limit;
	const paginatedComments = videoComments.slice(startIndex, endIndex);

	return {
		comments: paginatedComments,
		pagination: {
			page,
			limit,
			total: videoComments.length,
			totalPages: Math.ceil(videoComments.length / limit)
		}
	};
};

// ADMIN MOCK DATA HELPERS
export const getMockAdminVideos = (page = 1, limit = 20, filters: any = {}) => {
	const adminVideos = MOCK_VIDEOS.map(video => ({
		...video,
		status: Math.random() > 0.8 ? 'pending' : 'published',
		uploaded_by: {
			id: Math.floor(Math.random() * 100) + 1,
			name: `Dr. ${['John Smith', 'Sarah Johnson', 'Michael Brown', 'Rachel Davis'][Math.floor(Math.random() * 4)]}`,
			email: 'user@byu.edu'
		},
		upload_date: video.createdAt,
		views: video.viewCount,
		likes: video.likeCount,
		comments: Math.floor(Math.random() * 50),
		file_size: `${Math.floor(Math.random() * 500) + 100} MB`,
		resolution: '1080p',
		thumbnail: video.thumbnailUrl
	}));

	// Apply filters
	let filteredVideos = adminVideos;
	
	if (filters.search) {
		const searchLower = filters.search.toLowerCase();
		filteredVideos = filteredVideos.filter(video =>
			video.title.toLowerCase().includes(searchLower) ||
			video.description.toLowerCase().includes(searchLower)
		);
	}
	
	if (filters.category) {
		filteredVideos = filteredVideos.filter(video => 
			video.category.toLowerCase() === filters.category.toLowerCase()
		);
	}
	
	if (filters.status) {
		filteredVideos = filteredVideos.filter(video => video.status === filters.status);
	}

	// Pagination
	const startIndex = (page - 1) * limit;
	const endIndex = startIndex + limit;
	const paginatedVideos = filteredVideos.slice(startIndex, endIndex);

	return {
		videos: paginatedVideos,
		pagination: {
			page,
			limit,
			total: filteredVideos.length,
			totalPages: Math.ceil(filteredVideos.length / limit)
		}
	};
};

export const getMockVideoStats = () => {
	return {
		stats: {
			total_videos: MOCK_VIDEOS.length,
			published: Math.floor(MOCK_VIDEOS.length * 0.8),
			pending: Math.floor(MOCK_VIDEOS.length * 0.15),
			draft: Math.floor(MOCK_VIDEOS.length * 0.05),
			total_views: MOCK_VIDEOS.reduce((sum, video) => sum + video.viewCount, 0),
			total_likes: MOCK_VIDEOS.reduce((sum, video) => sum + video.likeCount, 0),
			total_comments: 1234,
			total_duration: "2847:32",
			storage_used: "45.6 GB",
			top_categories: MOCK_CATEGORIES.slice(0, 5).map(cat => ({
				name: cat.name,
				count: cat.videoCount,
				views: Math.floor(Math.random() * 10000) + 1000
			}))
		}
	};
};

// Mock data statistics
export const MOCK_DATA_STATISTICS = {
	totalVideos: 25,
	videoCategories: 12,
	mockComments: 15,
	totalArticles: 18,
	articleCategories: 8,
	articleTags: 25,
	authors: 6
};

// Article Categories
export const ARTICLE_CATEGORIES = [
	{ id: 1, name: 'Archaeological Evidence', slug: 'archaeology', description: 'Physical evidence and discoveries', count: 8 },
	{ id: 2, name: 'Historical Analysis', slug: 'history', description: 'Historical context and documentation', count: 6 },
	{ id: 3, name: 'Linguistic Studies', slug: 'linguistics', description: 'Language patterns and origins', count: 4 },
	{ id: 4, name: 'Geographic Research', slug: 'geography', description: 'Location studies and mapping', count: 5 },
	{ id: 5, name: 'Cultural Parallels', slug: 'culture', description: 'Cultural connections and similarities', count: 7 },
	{ id: 6, name: 'Scientific Analysis', slug: 'science', description: 'Scientific methods and findings', count: 3 },
	{ id: 7, name: 'Scholarly Reviews', slug: 'reviews', description: 'Academic reviews and critiques', count: 4 },
	{ id: 8, name: 'Recent Discoveries', slug: 'discoveries', description: 'Latest findings and research', count: 6 }
];

// Article Tags
export const ARTICLE_TAGS = [
	'Nephites', 'Lamanites', 'Ancient America', 'DNA Studies', 'Metallurgy', 'Agriculture',
	'Warfare', 'Trade Routes', 'Linguistics', 'Archaeology', 'Anthropology', 'Geography',
	'Mesoamerica', 'North America', 'Biblical Parallels', 'Ancient Texts', 'Chiasmus',
	'Hebrew Influences', 'Egyptian Connections', 'Pre-Columbian', 'Carbon Dating',
	'Excavations', 'Artifacts', 'Civilizations', 'Migration Patterns'
];

// Article Authors
export const ARTICLE_AUTHORS = [
	{
		id: 1,
		name: 'Dr. Sarah Mitchell',
		title: 'Professor of Archaeology',
		institution: 'Brigham Young University',
		bio: 'Leading expert in Mesoamerican archaeology with 20+ years of field experience.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 8
	},
	{
		id: 2,
		name: 'Dr. James Peterson',
		title: 'Biblical Scholar',
		institution: 'Harvard Divinity School',
		bio: 'Specialist in ancient Near Eastern texts and comparative religion.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 6
	},
	{
		id: 3,
		name: 'Dr. Maria Rodriguez',
		title: 'Linguistic Anthropologist',
		institution: 'Stanford University',
		bio: 'Expert in pre-Columbian languages and migration patterns.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 5
	},
	{
		id: 4,
		name: 'Dr. Robert Chen',
		title: 'Geneticist',
		institution: 'University of Utah',
		bio: 'Specialist in population genetics and ancient DNA analysis.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 4
	},
	{
		id: 5,
		name: 'Dr. Emily Thompson',
		title: 'Cultural Historian',
		institution: 'Yale University',
		bio: 'Expert in ancient American civilizations and cultural exchange.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 7
	},
	{
		id: 6,
		name: 'Dr. Michael Davis',
		title: 'Archaeological Scientist',
		institution: 'University of Pennsylvania',
		bio: 'Specialist in archaeological dating methods and scientific analysis.',
		avatar: '/api/placeholder/100/100',
		articlesCount: 3
	}
];

// Mock Articles Data
export const MOCK_ARTICLES = [
	{
		id: 1,
		title: 'Recent Archaeological Discoveries in Mesoamerica: New Evidence for Book of Mormon Civilizations',
		slug: 'mesoamerican-archaeological-discoveries-2024',
		excerpt: 'Groundbreaking excavations in Guatemala reveal complex urban centers that align with Book of Mormon descriptions of Nephite and Lamanite civilizations.',
		content: `Recent archaeological work in the Mirador Basin of Guatemala has uncovered evidence of sophisticated urban planning, advanced agricultural systems, and complex trade networks that bear striking similarities to civilizations described in the Book of Mormon.

The latest excavations, led by Dr. Richard Hansen, have revealed massive stone complexes dating to the Late Preclassic period (400 BC - 250 AD), precisely coinciding with the timeframe described in the Book of Mormon for major Nephite and Lamanite civilizations.

Key findings include:
- Advanced water management systems
- Evidence of large-scale agriculture supporting dense populations
- Sophisticated road networks connecting distant cities
- Defensive fortifications matching Book of Mormon descriptions
- Metallurgical evidence including copper and bronze artifacts

These discoveries continue to challenge previous assumptions about the complexity of ancient American civilizations and provide compelling archaeological context for Book of Mormon narratives.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 1,
		categoryId: 1,
		tags: ['Archaeology', 'Mesoamerica', 'Nephites', 'Lamanites', 'Excavations'],
		publishedAt: '2024-12-15T10:00:00Z',
		updatedAt: '2024-12-15T10:00:0Z',
		readTime: 8,
		views: 2847,
		likes: 156,
		featured: true,
		status: 'published'
	},
	{
		id: 2,
		title: 'DNA Studies and Ancient American Populations: Understanding Genetic Diversity',
		slug: 'dna-studies-ancient-american-populations',
		excerpt: 'New genetic research provides insights into the complex population history of ancient America and its implications for Book of Mormon studies.',
		content: `Recent advances in ancient DNA analysis have revolutionized our understanding of population movements and genetic diversity in pre-Columbian America. This research provides important context for discussions about Book of Mormon peoples and their potential genetic signatures.

Key research developments include:
- Improved extraction techniques for ancient DNA
- Broader sampling of ancient American populations
- Better understanding of population bottlenecks and founder effects
- Evidence for multiple migration events over thousands of years

Dr. Robert Chen's latest research demonstrates the complexity of ancient American genetics and challenges oversimplified models of population history. The evidence suggests a much more nuanced picture of ancient American peoples than previously understood.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 4,
		categoryId: 6,
		tags: ['DNA Studies', 'Ancient America', 'Migration Patterns', 'Genetics'],
		publishedAt: '2024-12-12T14:30:00Z',
		updatedAt: '2024-12-12T14:30:00Z',
		readTime: 12,
		views: 1923,
		likes: 89,
		featured: true,
		status: 'published'
	},
	{
		id: 3,
		title: 'Chiasmus in Ancient Literature: Patterns in the Book of Mormon',
		slug: 'chiasmus-ancient-literature-book-of-mormon',
		excerpt: 'Analysis of chiastic structures in the Book of Mormon reveals sophisticated literary patterns consistent with ancient Hebrew writing.',
		content: `Chiasmus, a literary device common in ancient Hebrew and Near Eastern texts, appears throughout the Book of Mormon in complex and sophisticated forms. This analysis examines the most compelling examples and their implications for Book of Mormon authorship.

Notable chiastic passages include:
- Alma 36: Alma's conversion narrative
- Mosiah 3-5: King Benjamin's speech
- Helaman 6: The cycle of righteousness and wickedness
- 3 Nephi 17: Christ's ministry among the Nephites

The complexity and consistency of these patterns suggest familiarity with ancient Hebrew literary traditions that would have been unknown to 19th-century Americans. Dr. James Peterson's comparative analysis with Dead Sea Scroll texts reveals remarkable parallels in structure and style.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 2,
		categoryId: 3,
		tags: ['Chiasmus', 'Hebrew Influences', 'Ancient Texts', 'Biblical Parallels'],
		publishedAt: '2024-12-10T09:15:00Z',
		updatedAt: '2024-12-10T09:15:00Z',
		readTime: 10,
		views: 1654,
		likes: 112,
		featured: false,
		status: 'published'
	},
	{
		id: 4,
		title: 'Ancient Trade Routes and the Book of Mormon: Economic Networks in Pre-Columbian America',
		slug: 'ancient-trade-routes-book-of-mormon-economic-networks',
		excerpt: 'Archaeological evidence reveals extensive trade networks in ancient America that align with Book of Mormon descriptions of economic activity.',
		content: `Recent research has mapped extensive trade networks throughout ancient America, revealing economic systems of remarkable sophistication. These networks distributed goods across thousands of miles and created the economic foundation for the large civilizations described in the Book of Mormon.

Evidence includes:
- Obsidian sourcing studies showing trade across Central America
- Cacao distribution networks extending from Mexico to South America
- Copper bell trade routes connecting distant regions
- Standardized weights and measures suggesting organized commerce

Dr. Emily Thompson's research demonstrates that these trade networks were not only extensive but also highly organized, requiring the kind of political and social structures described in Book of Mormon accounts of Nephite and Lamanite societies.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 5,
		categoryId: 4,
		tags: ['Trade Routes', 'Economics', 'Pre-Columbian', 'Civilizations'],
		publishedAt: '2024-12-08T16:45:00Z',
		updatedAt: '2024-12-08T16:45:00Z',
		readTime: 9,
		views: 1342,
		likes: 78,
		featured: false,
		status: 'published'
	},
	{
		id: 5,
		title: 'Metallurgy in Ancient America: Evidence for Advanced Technology',
		slug: 'metallurgy-ancient-america-advanced-technology',
		excerpt: 'Archaeological discoveries of sophisticated metalworking in pre-Columbian America support Book of Mormon descriptions of advanced technology.',
		content: `Contrary to popular misconceptions, advanced metallurgy was practiced throughout ancient America. Recent discoveries provide compelling evidence for the sophisticated metalworking described in the Book of Mormon.

Key discoveries include:
- Bronze artifacts from Poverty Point culture (1700-1100 BC)
- Copper mining operations in the Great Lakes region
- Advanced smelting techniques in South America
- Iron artifacts from pre-Columbian contexts

Dr. Michael Davis's analysis of these artifacts demonstrates technological capabilities that align with Book of Mormon descriptions of Nephite and Jaredite metallurgy. The evidence challenges assumptions about technological limitations in ancient America.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 6,
		categoryId: 6,
		tags: ['Metallurgy', 'Technology', 'Bronze', 'Copper', 'Ancient America'],
		publishedAt: '2024-12-05T11:20:00Z',
		updatedAt: '2024-12-05T11:20:00Z',
		readTime: 7,
		views: 1876,
		likes: 134,
		featured: false,
		status: 'published'
	},
	{
		id: 6,
		title: 'Linguistic Analysis: Hebrew and Egyptian Influences in Mesoamerican Languages',
		slug: 'linguistic-analysis-hebrew-egyptian-mesoamerican-languages',
		excerpt: 'Comparative linguistic studies reveal potential connections between Old World and New World languages that may support Book of Mormon claims.',
		content: `Dr. Maria Rodriguez's groundbreaking linguistic research has identified potential connections between ancient Hebrew and Egyptian languages and certain Mesoamerican language families. This research provides intriguing support for Book of Mormon claims about the linguistic heritage of its peoples.

Key findings include:
- Shared grammatical structures between Hebrew and Maya languages
- Potential Egyptian loanwords in Olmec-related languages
- Syntactic patterns suggesting ancient contact
- Phonological similarities in religious terminology

While the evidence is still being evaluated by the linguistic community, these findings suggest the possibility of ancient trans-oceanic contact that could explain the linguistic sophistication evident in the Book of Mormon text.`,
		featuredImage: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
		authorId: 3,
		categoryId: 3,
		tags: ['Linguistics', 'Hebrew Influences', 'Egyptian Connections', 'Mesoamerica'],
		publishedAt: '2024-12-03T13:30:00Z',
		updatedAt: '2024-12-03T13:30:00Z',
		readTime: 11,
		views: 1567,
		likes: 95,
		featured: false,
		status: 'published'
	}
	// ... Additional articles would continue here
];

// Helper function to get articles by category
export function getArticlesByCategory(categoryId: number) {
	return MOCK_ARTICLES.filter(article => article.categoryId === categoryId);
}

// Helper function to get articles by tag
export function getArticlesByTag(tag: string) {
	return MOCK_ARTICLES.filter(article => article.tags.includes(tag));
}

// Helper function to get articles by author
export function getArticlesByAuthor(authorId: number) {
	return MOCK_ARTICLES.filter(article => article.authorId === authorId);
}

// Helper function to search articles
export function searchArticles(query: string) {
	const searchTerm = query.toLowerCase();
	return MOCK_ARTICLES.filter(article => 
		article.title.toLowerCase().includes(searchTerm) ||
		article.excerpt.toLowerCase().includes(searchTerm) ||
		article.content.toLowerCase().includes(searchTerm) ||
		article.tags.some(tag => tag.toLowerCase().includes(searchTerm))
	);
}

// Helper function to get featured articles
export function getFeaturedArticles() {
	return MOCK_ARTICLES.filter(article => article.featured);
}

// Helper function to get recent articles
export function getRecentArticles(limit: number = 10) {
	return MOCK_ARTICLES
		.sort((a, b) => new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime())
		.slice(0, limit);
} 