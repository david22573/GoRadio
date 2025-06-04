export const load = async ({ fetch }) => {
	try {
		const res = await fetch('/api/stations');
		const data = await res.json();
		return { stations: data.stations };
	} catch (error) {
		console.error(error);
	}
};
