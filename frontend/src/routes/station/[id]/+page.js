export const load = async ({ fetch, params }) => {
	try {
		const res = await fetch('/api/stations');
		const data = await res.json();

		const station = data.stations?.find((s) => s.id === parseInt(params.id)) || null;

		return { station };
	} catch (error) {
		console.error(error);
		return { station: null };
	}
};
